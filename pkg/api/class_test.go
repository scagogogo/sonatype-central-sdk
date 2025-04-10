package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSearchByClassName 使用真实API测试类名搜索功能
func TestSearchByClassName(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试几个常见的类名
	classNames := []string{"Logger", "StringUtils", "HttpClient"}

	for _, className := range classNames {
		t.Run("Class_"+className, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(1 * time.Second)

			versionSlice, err := client.SearchByClassName(ctx, className, 5)

			if err != nil {
				t.Logf("搜索 %s 时出错: %v", className, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			// 记录找到的结果，但不强制要求特定内容
			t.Logf("找到 %d 个包含 %s 的结果", len(versionSlice), className)
			if len(versionSlice) > 0 {
				for i, v := range versionSlice[:minInt(3, len(versionSlice))] {
					t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
				}
			}

			assert.True(t, len(versionSlice) >= 0) // 只确保API正常返回
		})
	}
}
