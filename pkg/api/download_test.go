package api

import (
	"context"
	"testing"

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
