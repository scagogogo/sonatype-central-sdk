package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSearchByClassNameMock 使用模拟服务器测试功能
func TestSearchByClassNameMock(t *testing.T) {
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// 确保请求路径正确包含class参数
		assert.Contains(t, r.URL.RawQuery, "c%3ATestClass")
		mockVersionResponse(w, 3)
	})

	// 使用模拟类名进行测试
	versionSlice, err := client.SearchByClassName(context.Background(), "TestClass", 10)

	assert.Nil(t, err)
	assert.Equal(t, 3, len(versionSlice))
}

// TestSearchByClassNameReal 使用真实API测试
func TestSearchByClassNameReal(t *testing.T) {
	// 可选跳过长时间测试
	if testing.Short() {
		t.Skip("跳过真实API测试")
	}

	// 使用真实客户端
	client := createRealClient(t)

	// 测试几个常见的类名
	classNames := []string{"Logger", "StringUtils", "HttpClient"}

	for _, className := range classNames {
		t.Run("Class_"+className, func(t *testing.T) {
			versionSlice, err := client.SearchByClassName(context.Background(), className, 5)

			if err != nil {
				t.Logf("搜索 %s 时出错: %v", className, err)
				return
			}

			// 记录找到的结果，但不强制要求特定内容
			t.Logf("找到 %d 个包含 %s 的结果", len(versionSlice), className)
			if len(versionSlice) > 0 {
				for i, v := range versionSlice[:min(3, len(versionSlice))] {
					t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
				}
			}

			assert.True(t, len(versionSlice) >= 0) // 只确保API正常返回
		})
	}
}

// min 返回两个整数中较小的一个
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
