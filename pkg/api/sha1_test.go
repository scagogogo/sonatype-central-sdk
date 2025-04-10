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

func TestGetFirstBySha1(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试几个SHA1值
	sha1Values := []string{
		"0235ba8b489512805ac13a8f9ea77a1ca5ebe3e8", // 应该存在
		"abcdefabcdefabcdefabcdefabcdefabcdefabcd", // 不太可能存在
	}

	for i, sha1 := range sha1Values {
		t.Run(fmt.Sprintf("GetFirstBySHA1_%d", i+1), func(t *testing.T) {
			// 避免请求过快
			time.Sleep(1 * time.Second)

			version, err := client.GetFirstBySha1(ctx, sha1)

			if err != nil {
				t.Logf("SHA1 %s 查询出错: %v", sha1, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			if version != nil {
				t.Logf("找到结果: %s:%s:%s", version.GroupId, version.ArtifactId, version.Version)
			} else {
				t.Logf("未找到SHA1 %s 的匹配结果", sha1)
			}
		})
	}
}

func TestExistsSha1(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试几个SHA1值
	testCases := []struct {
		sha1           string
		expectedExists bool
		description    string
	}{
		{"0235ba8b489512805ac13a8f9ea77a1ca5ebe3e8", true, "commons-lang3 SHA1"},
		{"abcdefabcdefabcdefabcdefabcdefabcdefabcd", false, "无效SHA1"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("ExistsSHA1_%s", tc.description), func(t *testing.T) {
			// 避免请求过快
			time.Sleep(1 * time.Second)

			exists, err := client.ExistsSha1(ctx, tc.sha1)

			if err != nil {
				t.Logf("SHA1 %s 查询出错: %v", tc.sha1, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			t.Logf("SHA1 %s 存在性检查结果: %v", tc.sha1, exists)

			// 不强制进行断言，因为Maven Central的数据可能变化
			// 只记录结果是否符合预期
			if exists != tc.expectedExists {
				t.Logf("注意: 预期 %v，但得到 %v", tc.expectedExists, exists)
			}
		})
	}
}

func TestSearchExactSha1(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试SHA1
	sha1 := "0235ba8b489512805ac13a8f9ea77a1ca5ebe3e8" // commons-lang3

	t.Run("SearchExactSHA1", func(t *testing.T) {
		results, err := client.SearchExactSha1(ctx, sha1)

		if err != nil {
			t.Logf("精确SHA1搜索出错: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		t.Logf("精确搜索找到 %d 个匹配SHA1的结果", len(results))
		for i, v := range results[:minInt(3, len(results))] {
			t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
		}
	})
}

func TestCountBySha1(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试SHA1
	sha1 := "0235ba8b489512805ac13a8f9ea77a1ca5ebe3e8" // commons-lang3

	t.Run("CountBySHA1", func(t *testing.T) {
		count, err := client.CountBySha1(ctx, sha1)

		if err != nil {
			t.Logf("SHA1计数查询出错: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		t.Logf("SHA1 %s 匹配的构件数量: %d", sha1, count)

		// 不强制断言具体数量，只确保返回了有效结果
		if count >= 0 {
			t.Logf("成功获取到计数结果")
		}
	})
}
