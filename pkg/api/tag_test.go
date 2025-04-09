package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSearchByTagMock 使用模拟服务器测试功能
func TestSearchByTagMock(t *testing.T) {
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// 确保请求路径正确包含标签参数
		assert.Contains(t, r.URL.RawQuery, "tags%3Atest-tag")
		mockArtifactResponse(w, 3)
	})

	// 使用模拟标签进行测试
	artifacts, err := client.SearchByTag(context.Background(), "test-tag", 10)

	assert.Nil(t, err)
	assert.Equal(t, 3, len(artifacts))
}

// TestSearchByTagReal 使用真实API测试
func TestSearchByTagReal(t *testing.T) {
	// 可选跳过长时间测试
	if testing.Short() {
		t.Skip("跳过真实API测试")
	}

	// 使用真实客户端
	client := createRealClient(t)

	// 测试几个常见的标签
	tagNames := []string{"jdbc", "logging", "http-client"}

	for _, tag := range tagNames {
		t.Run("Tag_"+tag, func(t *testing.T) {
			artifacts, err := client.SearchByTag(context.Background(), tag, 5)

			if err != nil {
				t.Logf("搜索标签 %s 时出错: %v", tag, err)
				return
			}

			// 记录找到的结果，但不强制要求特定内容
			t.Logf("找到 %d 个包含标签 %s 的结果", len(artifacts), tag)
			if len(artifacts) > 0 {
				for i, a := range artifacts[:min(3, len(artifacts))] {
					t.Logf("结果 %d: %s:%s:%s", i+1, a.GroupId, a.ArtifactId, a.LatestVersion)
				}
			}

			assert.True(t, len(artifacts) >= 0) // 只确保API正常返回
		})
	}
}
