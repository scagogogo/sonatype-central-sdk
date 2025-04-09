package api

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
