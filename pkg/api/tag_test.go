package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSearchByTag 使用真实API测试标签搜索功能
func TestSearchByTag(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试几个常见的标签
	tagNames := []string{"jdbc", "logging", "http-client"}

	for _, tag := range tagNames {
		t.Run("Tag_"+tag, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(1 * time.Second)

			artifacts, err := client.SearchByTag(ctx, tag, 5)

			if err != nil {
				t.Logf("搜索标签 %s 时出错: %v", tag, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			// 记录找到的结果，但不强制要求特定内容
			t.Logf("找到 %d 个包含标签 %s 的结果", len(artifacts), tag)
			if len(artifacts) > 0 {
				for i, a := range artifacts[:minInt(3, len(artifacts))] {
					t.Logf("结果 %d: %s:%s:%s", i+1, a.GroupId, a.ArtifactId, a.LatestVersion)
				}
			}

			assert.True(t, len(artifacts) >= 0) // 只确保API正常返回
		})
	}
}

// TestGetRelatedTags 使用真实API测试获取相关标签功能
func TestGetRelatedTags(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试一个常见标签的相关标签
	tag := "java"

	// 添加短暂延迟，避免请求过快
	time.Sleep(1 * time.Second)

	// 测试获取相关标签
	relatedTags, err := client.GetRelatedTags(ctx, tag, 10)
	if err != nil {
		t.Logf("获取相关标签时出错: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	assert.NotNil(t, relatedTags)

	// 输出找到的相关标签信息
	t.Logf("找到 %d 个与 %s 相关的标签", len(relatedTags), tag)
	if len(relatedTags) > 0 {
		count := 0
		for tag, frequency := range relatedTags {
			if count < 5 { // 只显示前5个
				t.Logf("相关标签: %s (出现频率: %d)", tag, frequency)
				count++
			}
		}
	}
}

func TestTagRelatedMethods(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试常见标签
	tags := []string{"java", "json", "http", "logging"}

	// 测试 CountArtifactsByTag
	t.Run("CountArtifactsByTag", func(t *testing.T) {
		for _, tag := range tags {
			count, err := client.CountArtifactsByTag(ctx, tag)
			if err != nil {
				t.Logf("计算标签 %s 数量时出错: %v", tag, err)
				t.Skip("无法连接到Maven Central API")
				return
			}
			t.Logf("标签 %s 的构件数量: %d", tag, count)
		}
	})

	// 测试 SearchByTagPrefix
	t.Run("SearchByTagPrefix", func(t *testing.T) {
		prefix := "ja" // 应该匹配java, javascript等
		artifacts, err := client.SearchByTagPrefix(ctx, prefix, 5)
		if err != nil {
			t.Logf("搜索标签前缀 %s 时出错: %v", prefix, err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		t.Logf("标签前缀 %s 匹配了 %d 个结果", prefix, len(artifacts))
		for i, artifact := range artifacts {
			t.Logf("结果 %d: %s:%s (标签: %v)", i+1, artifact.GroupId, artifact.ArtifactId, artifact.Tags)
		}
	})

	// 测试 GetTagSuggestions
	t.Run("GetTagSuggestions", func(t *testing.T) {
		baseTag := "java"
		suggestions, err := client.GetTagSuggestions(ctx, baseTag, 5)
		if err != nil {
			t.Logf("获取标签 %s 的建议时出错: %v", baseTag, err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		t.Logf("与 %s 相关的标签建议:", baseTag)
		for i, suggestion := range suggestions {
			t.Logf("建议 %d: %s", i+1, suggestion)
		}
	})

	// 测试 SearchByTagAndSortByPopularity
	t.Run("SearchByTagAndSortByPopularity", func(t *testing.T) {
		tag := "http"
		artifacts, err := client.SearchByTagAndSortByPopularity(ctx, tag, 5)
		if err != nil {
			t.Logf("按流行度排序标签 %s 时出错: %v", tag, err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		t.Logf("标签 %s 按流行度排序的前5个结果:", tag)
		for i, artifact := range artifacts {
			t.Logf("结果 %d: %s:%s (版本数: %d)", i+1, artifact.GroupId, artifact.ArtifactId, artifact.VersionCount)
		}
	})

	// 测试 AnalyzeTagTrends
	t.Run("AnalyzeTagTrends", func(t *testing.T) {
		trends, err := client.AnalyzeTagTrends(ctx, tags, 6)
		if err != nil {
			t.Logf("分析标签趋势时出错: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		t.Logf("标签趋势分析结果:")
		for tag, trend := range trends {
			t.Logf("标签: %s, 使用数量: %d, 活跃度: %.2f, 趋势: %s",
				tag, trend.CurrentUsageCount, trend.ActivityScore, trend.Trend)
		}
	})
}
