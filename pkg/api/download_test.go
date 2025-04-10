package api

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDownload(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 测试真实API - 下载一个常见的POM文件
	filePath := "com/google/inject/guice/4.0/guice-4.0.pom"
	download, err := client.Download(context.Background(), filePath)

	assert.Nil(t, err)
	assert.NotEmpty(t, download)

	// 验证下载的POM文件内容
	pomContent := string(download)
	assert.Contains(t, pomContent, "<artifactId>guice</artifactId>")
	assert.Contains(t, pomContent, "<version>4.0</version>")
}

func TestDownloadMultipleFiles(t *testing.T) {
	client := createRealClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试下载多个不同类型的文件
	groupId := "com.google.inject"
	artifactId := "guice"
	version := "4.0"

	fileTypes := []ArtifactFile{
		PomFile,
		JarFile,
		SourcesFile,
	}

	results := client.DownloadMultipleFiles(ctx, groupId, artifactId, version, fileTypes)

	// 验证结果
	assert.Equal(t, 3, len(results), "应该下载3种类型的文件")

	// 检查POM文件
	pomResult := results["POM"]
	assert.NotNil(t, pomResult)
	assert.Nil(t, pomResult.Error)
	assert.NotEmpty(t, pomResult.Data)

	// 检查JAR文件
	jarResult := results["JAR"]
	assert.NotNil(t, jarResult)
	assert.Nil(t, jarResult.Error)
	assert.NotEmpty(t, jarResult.Data)
}

func TestDownloadWithChecksum(t *testing.T) {
	client := createRealClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试下载并验证SHA1校验和
	filePath := "com/google/inject/guice/4.0/guice-4.0.pom"

	data, checksum, err := client.DownloadWithChecksum(ctx, filePath, "sha1")

	// 验证结果
	assert.Nil(t, err)
	assert.NotEmpty(t, data)
	assert.NotEmpty(t, checksum)
	assert.Len(t, checksum, 40) // SHA1长度为40个字符
}

func TestDownloadAsync(t *testing.T) {
	client := createRealClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试异步下载
	filePath := "com/google/inject/guice/4.0/guice-4.0.pom"

	resultChan := client.DownloadAsync(ctx, filePath)

	// 等待结果
	result := <-resultChan

	// 验证结果
	assert.Nil(t, result.Error)
	assert.NotEmpty(t, result.Data)
}

func TestDownloadCompleteBundle(t *testing.T) {
	client := createRealClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试下载完整包
	groupId := "com.google.inject"
	artifactId := "guice"
	version := "4.0"

	bundle, err := client.DownloadCompleteBundle(ctx, groupId, artifactId, version)

	// 验证结果
	assert.Nil(t, err)
	assert.NotNil(t, bundle)
	assert.Equal(t, groupId, bundle.GroupId)
	assert.Equal(t, artifactId, bundle.ArtifactId)
	assert.Equal(t, version, bundle.Version)

	// 至少POM文件应该存在
	assert.NotEmpty(t, bundle.Pom)
}

func TestSaveBundle(t *testing.T) {
	client := createRealClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试下载并保存完整包
	groupId := "com.google.inject"
	artifactId := "guice"
	version := "4.0"

	bundle, err := client.DownloadCompleteBundle(ctx, groupId, artifactId, version)
	assert.Nil(t, err)

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "maven-bundle-test")
	assert.Nil(t, err)
	defer os.RemoveAll(tempDir) // 测试结束后清理

	// 保存包
	err = client.SaveBundle(bundle, tempDir)
	assert.Nil(t, err)

	// 验证文件是否被正确保存
	pomPath := filepath.Join(tempDir,
		strings.ReplaceAll(groupId, ".", string(filepath.Separator)),
		artifactId,
		version,
		artifactId+"-"+version+".pom")

	_, err = os.Stat(pomPath)
	assert.Nil(t, err, "POM文件应该存在于指定路径")
}
