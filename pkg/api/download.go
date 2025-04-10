package api

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// 定义常用的文件扩展名
const (
	POM     = "pom"
	JAR     = "jar"
	WAR     = "war"
	AAR     = "aar"
	SOURCES = "sources.jar"
	JAVADOC = "javadoc.jar"
	TESTS   = "tests.jar"
)

// Download 从Maven中央仓库下载文件
// filepath比如： com/jolira/guice/3.0.0/guice-3.0.0.pom
func (c *Client) Download(ctx context.Context, filePath string) ([]byte, error) {
	return c.downloadWithCache(ctx, filePath)
}

// DownloadFile 下载文件并保存到本地路径
func (c *Client) DownloadFile(ctx context.Context, filePath, localPath string) error {
	data, err := c.Download(ctx, filePath)
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(localPath, data, 0644)
}

// DownloadToWriter 下载文件并写入指定的writer
func (c *Client) DownloadToWriter(ctx context.Context, filePath string, writer io.Writer) error {
	data, err := c.Download(ctx, filePath)
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	return err
}

// BuildArtifactPath 构建制品路径
func BuildArtifactPath(groupId, artifactId, version, extension string, classifier ...string) string {
	basePath := fmt.Sprintf("%s/%s/%s", strings.ReplaceAll(groupId, ".", "/"), artifactId, version)

	fileName := fmt.Sprintf("%s-%s", artifactId, version)

	// 添加分类器
	if len(classifier) > 0 && classifier[0] != "" {
		fileName += "-" + classifier[0]
	}

	// 添加扩展名
	fileName += "." + extension

	return fmt.Sprintf("%s/%s", basePath, fileName)
}

// DownloadPom 下载POM文件
func (c *Client) DownloadPom(ctx context.Context, groupId, artifactId, version string) ([]byte, error) {
	path := BuildArtifactPath(groupId, artifactId, version, POM)
	return c.Download(ctx, path)
}

// DownloadJar 下载JAR文件
func (c *Client) DownloadJar(ctx context.Context, groupId, artifactId, version string) ([]byte, error) {
	path := BuildArtifactPath(groupId, artifactId, version, JAR)
	return c.Download(ctx, path)
}

// DownloadSources 下载源码JAR
func (c *Client) DownloadSources(ctx context.Context, groupId, artifactId, version string) ([]byte, error) {
	path := BuildArtifactPath(groupId, artifactId, version, JAR, "sources")
	return c.Download(ctx, path)
}

// DownloadJavadoc 下载JavaDoc
func (c *Client) DownloadJavadoc(ctx context.Context, groupId, artifactId, version string) ([]byte, error) {
	path := BuildArtifactPath(groupId, artifactId, version, JAR, "javadoc")
	return c.Download(ctx, path)
}

// DownloadArtifact 根据指定参数下载制品
func (c *Client) DownloadArtifact(ctx context.Context, artifact *response.Artifact, extension string, classifier ...string) ([]byte, error) {
	path := BuildArtifactPath(artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion, extension, classifier...)
	return c.Download(ctx, path)
}

// DownloadArtifactWithVersion 根据指定版本下载制品
func (c *Client) DownloadArtifactWithVersion(ctx context.Context, artifact *response.Version, extension string, classifier ...string) ([]byte, error) {
	path := BuildArtifactPath(artifact.GroupId, artifact.ArtifactId, artifact.Version, extension, classifier...)
	return c.Download(ctx, path)
}

// ArtifactFile 表示一个制品文件类型
type ArtifactFile struct {
	Type       string // 文件类型的标识，如"pom", "jar"等
	Extension  string // 文件扩展名
	Classifier string // 可选的分类器
}

// 预定义的常用制品文件类型
var (
	PomFile     = ArtifactFile{Type: "POM", Extension: POM}
	JarFile     = ArtifactFile{Type: "JAR", Extension: JAR}
	SourcesFile = ArtifactFile{Type: "SOURCES", Extension: JAR, Classifier: "sources"}
	JavadocFile = ArtifactFile{Type: "JAVADOC", Extension: JAR, Classifier: "javadoc"}
	TestsFile   = ArtifactFile{Type: "TESTS", Extension: JAR, Classifier: "tests"}
	WarFile     = ArtifactFile{Type: "WAR", Extension: WAR}
	AarFile     = ArtifactFile{Type: "AAR", Extension: AAR}
)

// CommonArtifactFiles 返回常用的制品文件类型列表
func CommonArtifactFiles() []ArtifactFile {
	return []ArtifactFile{
		PomFile, JarFile, SourcesFile, JavadocFile,
	}
}

// DownloadResult 表示一个下载结果
type DownloadResult struct {
	FileType ArtifactFile // 文件类型
	Data     []byte       // 文件数据
	Error    error        // 下载过程中的错误，如果有的话
	Path     string       // 文件在仓库中的路径
	SHA1     string       // SHA1摘要，仅当验证时可用
	MD5      string       // MD5摘要，仅当验证时可用
	SHA256   string       // SHA256摘要，仅当验证时可用
}

// DownloadMultipleFiles 下载多个不同类型的文件
// 参数:
//   - ctx: 上下文
//   - groupId: 组ID
//   - artifactId: 制品ID
//   - version: 版本号
//   - fileTypes: 要下载的文件类型列表
//
// 返回:
//   - map[string]*DownloadResult: 以文件类型标识为键的下载结果映射
func (c *Client) DownloadMultipleFiles(ctx context.Context, groupId, artifactId, version string, fileTypes []ArtifactFile) map[string]*DownloadResult {
	results := make(map[string]*DownloadResult)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, fileType := range fileTypes {
		wg.Add(1)
		go func(ft ArtifactFile) {
			defer wg.Done()

			path := BuildArtifactPath(groupId, artifactId, version, ft.Extension, ft.Classifier)
			data, err := c.Download(ctx, path)

			result := &DownloadResult{
				FileType: ft,
				Data:     data,
				Error:    err,
				Path:     path,
			}

			mu.Lock()
			results[ft.Type] = result
			mu.Unlock()
		}(fileType)
	}

	wg.Wait()
	return results
}

// DownloadWithChecksum 下载文件并验证其校验和
// 参数:
//   - ctx: 上下文
//   - filePath: 文件路径
//   - checksumType: 校验和类型，支持"sha1", "md5", "sha256"
//
// 返回:
//   - []byte: 文件内容
//   - string: 计算出的校验和
//   - error: 如果发生错误
func (c *Client) DownloadWithChecksum(ctx context.Context, filePath string, checksumType string) ([]byte, string, error) {
	// 下载文件
	data, err := c.Download(ctx, filePath)
	if err != nil {
		return nil, "", err
	}

	// 计算校验和
	var checksum string
	switch strings.ToLower(checksumType) {
	case "sha1":
		hash := sha1.Sum(data)
		checksum = hex.EncodeToString(hash[:])
	case "md5":
		hash := md5.Sum(data)
		checksum = hex.EncodeToString(hash[:])
	case "sha256":
		hash := sha256.Sum256(data)
		checksum = hex.EncodeToString(hash[:])
	default:
		return data, "", fmt.Errorf("不支持的校验和类型: %s", checksumType)
	}

	// 下载对应的校验和文件进行验证
	checksumFilePath := filePath + "." + checksumType
	checksumFileData, err := c.Download(ctx, checksumFilePath)

	// 如果校验和文件不存在，直接返回计算出的校验和
	if err != nil {
		return data, checksum, nil
	}

	// 校验和文件通常只包含十六进制字符串
	remoteChecksum := strings.TrimSpace(string(checksumFileData))

	// 有时校验和文件可能包含文件名，需要提取第一段
	if parts := strings.Fields(remoteChecksum); len(parts) > 0 {
		remoteChecksum = parts[0]
	}

	// 比较校验和
	if checksum != remoteChecksum {
		return data, checksum, fmt.Errorf("校验和不匹配: 计算得到 %s，远程值 %s", checksum, remoteChecksum)
	}

	return data, checksum, nil
}

// DownloadProgress 下载进度回调函数
type DownloadProgress func(downloaded, total int64, fileName string)

// AsyncDownloadResult 异步下载结果
type AsyncDownloadResult struct {
	FileType ArtifactFile
	Data     []byte
	Error    error
	Path     string
}

// DownloadAsync 异步下载文件
// 返回一个通道，当下载完成时会发送结果
func (c *Client) DownloadAsync(ctx context.Context, filePath string) <-chan AsyncDownloadResult {
	resultChan := make(chan AsyncDownloadResult, 1)

	go func() {
		defer close(resultChan)

		data, err := c.Download(ctx, filePath)
		result := AsyncDownloadResult{
			Data:  data,
			Error: err,
			Path:  filePath,
		}

		resultChan <- result
	}()

	return resultChan
}

// ArtifactBundle 表示一个制品包，包含所有相关文件
type ArtifactBundle struct {
	GroupId    string
	ArtifactId string
	Version    string
	Pom        []byte
	Jar        []byte
	Sources    []byte
	Javadoc    []byte
	Tests      []byte
	OtherFiles map[string][]byte
	Errors     map[string]error
}

// DownloadCompleteBundle 下载制品的完整包，包括所有可用的相关文件
// 参数:
//   - ctx: 上下文
//   - groupId: 组ID
//   - artifactId: 制品ID
//   - version: 版本号
//   - extraFiles: 额外要下载的文件类型列表
//
// 返回:
//   - *ArtifactBundle: 包含所有文件的包
//   - error: 如果所有必要文件都下载失败
func (c *Client) DownloadCompleteBundle(ctx context.Context, groupId, artifactId, version string, extraFiles ...ArtifactFile) (*ArtifactBundle, error) {
	// 创建基本的bundle结构
	bundle := &ArtifactBundle{
		GroupId:    groupId,
		ArtifactId: artifactId,
		Version:    version,
		OtherFiles: make(map[string][]byte),
		Errors:     make(map[string]error),
	}

	// 必要文件列表
	essentialFiles := []struct {
		fileType ArtifactFile
		target   *[]byte
		name     string
	}{
		{PomFile, &bundle.Pom, "POM"},
		{JarFile, &bundle.Jar, "JAR"},
		{SourcesFile, &bundle.Sources, "SOURCES"},
		{JavadocFile, &bundle.Javadoc, "JAVADOC"},
		{TestsFile, &bundle.Tests, "TESTS"},
	}

	// 下载必要文件
	var wg sync.WaitGroup
	var mu sync.Mutex
	var essentialErrors int = 0

	for _, file := range essentialFiles {
		wg.Add(1)
		go func(ft ArtifactFile, target *[]byte, name string) {
			defer wg.Done()

			path := BuildArtifactPath(groupId, artifactId, version, ft.Extension, ft.Classifier)
			data, err := c.Download(ctx, path)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				bundle.Errors[name] = err
				if name == "POM" || name == "JAR" {
					essentialErrors++
				}
			} else {
				*target = data
			}
		}(file.fileType, file.target, file.name)
	}

	// 下载额外文件
	for _, extraFile := range extraFiles {
		wg.Add(1)
		go func(ft ArtifactFile) {
			defer wg.Done()

			path := BuildArtifactPath(groupId, artifactId, version, ft.Extension, ft.Classifier)
			data, err := c.Download(ctx, path)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				bundle.Errors[ft.Type] = err
			} else {
				bundle.OtherFiles[ft.Type] = data
			}
		}(extraFile)
	}

	wg.Wait()

	// 检查是否所有必要文件都下载失败
	if essentialErrors >= 2 {
		return bundle, errors.New("必要文件（POM和JAR）下载失败")
	}

	return bundle, nil
}

// SaveBundle 将制品包保存到本地目录
// 参数:
//   - bundle: 要保存的制品包
//   - baseDir: 基础目录
//
// 返回:
//   - error: 如果保存过程中发生错误
func (c *Client) SaveBundle(bundle *ArtifactBundle, baseDir string) error {
	// 构建目标目录
	targetDir := filepath.Join(baseDir,
		strings.ReplaceAll(bundle.GroupId, ".", string(filepath.Separator)),
		bundle.ArtifactId,
		bundle.Version)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	// 保存文件
	filesToSave := []struct {
		data     []byte
		fileName string
	}{
		{bundle.Pom, fmt.Sprintf("%s-%s.pom", bundle.ArtifactId, bundle.Version)},
		{bundle.Jar, fmt.Sprintf("%s-%s.jar", bundle.ArtifactId, bundle.Version)},
		{bundle.Sources, fmt.Sprintf("%s-%s-sources.jar", bundle.ArtifactId, bundle.Version)},
		{bundle.Javadoc, fmt.Sprintf("%s-%s-javadoc.jar", bundle.ArtifactId, bundle.Version)},
		{bundle.Tests, fmt.Sprintf("%s-%s-tests.jar", bundle.ArtifactId, bundle.Version)},
	}

	for _, file := range filesToSave {
		if file.data == nil || len(file.data) == 0 {
			continue
		}

		filePath := filepath.Join(targetDir, file.fileName)
		if err := os.WriteFile(filePath, file.data, 0644); err != nil {
			return err
		}
	}

	// 保存其他文件
	for fileType, data := range bundle.OtherFiles {
		fileName := fmt.Sprintf("%s-%s-%s.%s",
			bundle.ArtifactId, bundle.Version, strings.ToLower(fileType), "jar")
		filePath := filepath.Join(targetDir, fileName)

		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}
