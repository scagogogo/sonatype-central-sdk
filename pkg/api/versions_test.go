package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchByGroupIdAndArtifactId(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 测试真实API - 使用常见的依赖项进行测试
	versionSlice, err := client.ListVersions(context.Background(), "com.google.inject", "guice", -1)
	assert.Nil(t, err)
	assert.True(t, len(versionSlice) > 0)

	// 验证至少返回了几个常见版本
	var hasVersion4 bool
	for _, v := range versionSlice {
		if v.Version == "4.0" {
			hasVersion4 = true
			break
		}
	}
	assert.True(t, hasVersion4, "应该包含4.0版本")
}
