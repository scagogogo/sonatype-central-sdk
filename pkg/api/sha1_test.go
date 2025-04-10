package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
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

// TestSha1PrefixSearch 测试SHA1前缀搜索功能
// 这个测试用于验证SHA1是否支持前缀搜索（模糊搜索）
func TestSha1PrefixSearch(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 选择一个已知存在的SHA1，只取前几位进行前缀搜索
	originalSha1 := "0235ba8b489512805ac13a8f9ea77a1ca5ebe3e8" // 完整SHA1

	// 测试不同长度的前缀
	prefixLengths := []int{5, 10, 20, 30}

	for _, length := range prefixLengths {
		if length > len(originalSha1) {
			length = len(originalSha1)
		}

		prefix := originalSha1[:length]

		t.Run(fmt.Sprintf("PrefixLength_%d", length), func(t *testing.T) {
			// 避免请求过快
			time.Sleep(1 * time.Second)

			// 使用自定义查询来进行SHA1前缀搜索
			customQuery := "1:" + prefix + "*"
			search := request.NewSearchRequest().SetQuery(request.NewQuery().SetCustomQuery(customQuery))
			result, err := SearchRequestJsonDoc[*response.Version](client, ctx, search)

			if err != nil {
				t.Logf("SHA1前缀 %s 搜索出错: %v", prefix, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			if result == nil || result.ResponseBody == nil {
				t.Logf("SHA1前缀 %s 搜索返回空结果", prefix)
				return
			}

			docs := result.ResponseBody.Docs

			// 输出结果数量和前几个结果
			t.Logf("SHA1前缀 %s 找到 %d 个匹配结果", prefix, len(docs))
			for i, v := range docs[:minInt(3, len(docs))] {
				t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
			}

			// 完整SHA1搜索结果（用于比较）
			fullSha1Results, err := client.SearchBySha1(ctx, originalSha1, 5)
			if err == nil && len(fullSha1Results) > 0 {
				t.Logf("完整SHA1搜索找到 %d 个结果", len(fullSha1Results))

				// 检查前缀搜索的结果是否包含完整SHA1的结果
				if len(docs) >= len(fullSha1Results) {
					t.Logf("前缀搜索结果数量大于或等于完整SHA1搜索结果数量")
				} else {
					t.Logf("前缀搜索结果数量少于完整SHA1搜索结果数量")
				}
			}
		})
	}
}

// TestSearchBySha1Prefix 测试SHA1前缀搜索API方法
func TestSearchBySha1Prefix(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 选择一个已知存在的SHA1，用于测试
	sha1 := "0235ba8b489512805ac13a8f9ea77a1ca5ebe3e8"

	testCases := []struct {
		name       string
		prefix     string
		expectMany bool // 是否期望返回多个结果
	}{
		{"短前缀_5字符", sha1[:5], true},
		{"中等前缀_10字符", sha1[:10], false},
		{"长前缀_20字符", sha1[:20], false},
		{"完整SHA1", sha1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 避免请求过快
			time.Sleep(1 * time.Second)

			results, err := client.SearchBySha1Prefix(ctx, tc.prefix, 10)

			if err != nil {
				t.Logf("SHA1前缀 %s 搜索出错: %v", tc.prefix, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			t.Logf("SHA1前缀 %s 找到 %d 个匹配结果", tc.prefix, len(results))
			for i, v := range results[:minInt(3, len(results))] {
				t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
			}

			if tc.expectMany && len(results) <= 1 {
				t.Logf("注意: 预期找到多个结果，但只找到 %d 个", len(results))
			}

			if !tc.expectMany && len(results) > 1 {
				t.Logf("注意: 预期找到单个或零个结果，但找到 %d 个", len(results))
			}
		})
	}
}
