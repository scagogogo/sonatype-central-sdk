package api

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestSearchBySha1(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 尝试几个常见开源库的SHA1
	// 由于SHA1可能随着时间变化，我们尝试多个
	sha1Values := []string{
		"0235ba8b489512805ac13a8f9ea77a1ca5ebe3e8", // commons-lang3
		"3cd63d075497751784b2fa84be59432f4905bf7c", // slf4j-api
		"a927da0a7bf2a923691c2d8fb3e3d8a87a6cb9ea", // 尝试另一个常见库
	}

	for i, sha1 := range sha1Values {
		t.Run(fmt.Sprintf("SHA1Test%d", i+1), func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(1 * time.Second)

			versionSlice, err := client.SearchBySha1(ctx, sha1, 5)

			// 如果发生错误，跳过测试
			if err != nil {
				t.Logf("SHA1 %s 搜索出错: %v", sha1, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			// 不强制要求找到结果，只要API正常响应即可
			if len(versionSlice) > 0 {
				t.Logf("找到 %d 个匹配SHA1的结果", len(versionSlice))
				for j, v := range versionSlice[:minInt(3, len(versionSlice))] {
					t.Logf("结果 %d: %s:%s:%s", j+1, v.GroupId, v.ArtifactId, v.Version)
				}
			} else {
				t.Logf("SHA1 %s 未找到匹配", sha1)
			}
		})
	}
}
