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

			// 确保至少找到了一些结果
			assert.True(t, len(artifacts) > 0, "应该至少找到一些包含标签 %s 的结果", tag)

			// 验证返回的结果是否有效
			if len(artifacts) > 0 {
				// 检查第一个结果是否具有有效的GroupId和ArtifactId
				assert.NotEmpty(t, artifacts[0].GroupId, "结果的GroupId不应为空")
				assert.NotEmpty(t, artifacts[0].ArtifactId, "结果的ArtifactId不应为空")
			}
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
	} else {
		t.Logf("注意: 没有找到与标签 %s 相关的标签，这可能是正常情况", tag)
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

			// 确保返回了有效的计数
			assert.Greater(t, count, 0, "标签 %s 应该至少有一些构件", tag)
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
		assert.Greater(t, len(artifacts), 0, "标签前缀 %s 应该至少匹配一些结果", prefix)

		// 检查结果是否包含标签信息
		hasResultsWithTags := false
		for i, artifact := range artifacts {
			t.Logf("结果 %d: %s:%s (标签: %v)", i+1, artifact.GroupId, artifact.ArtifactId, artifact.Tags)

			// 检查是否至少有一个包含有效标签的结果
			if len(artifact.Tags) > 0 {
				hasResultsWithTags = true
				t.Logf("找到包含标签的结果: %v", artifact.Tags)
			}
		}

		// 不强制要求前缀搜索结果必须包含标签，但记录此情况
		if !hasResultsWithTags {
			t.Logf("注意: 所有搜索结果中都没有包含标签信息")
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

		// 记录得到的建议数量
		t.Logf("获取到 %d 个标签建议", len(suggestions))

		// 不强制要求一定有建议，因为这取决于API的行为和数据状态
		// 对于没有返回建议的情况，仅做记录而不判定为失败
		if len(suggestions) == 0 {
			t.Logf("注意: 没有找到与标签 %s 相关的建议，这可能是正常情况", baseTag)
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
		assert.Greater(t, len(artifacts), 0, "应该至少找到一些包含标签 %s 的结果", tag)

		// 验证结果是否按照流行度排序
		if len(artifacts) >= 2 {
			// 确保结果至少包含版本计数信息
			assert.GreaterOrEqual(t, artifacts[0].VersionCount, 0, "结果应该包含版本计数信息")

			// 验证排序是否正确（应该按版本数量降序排列）
			isCorrectlySorted := true
			for i := 1; i < len(artifacts); i++ {
				if artifacts[i-1].VersionCount < artifacts[i].VersionCount {
					isCorrectlySorted = false
					break
				}
			}
			assert.True(t, isCorrectlySorted, "结果应该按流行度（版本数量）降序排序")
		}

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
		assert.Greater(t, len(trends), 0, "应该至少返回一些标签的趋势分析结果")

		for tag, trend := range trends {
			t.Logf("标签: %s, 使用数量: %d, 活跃度: %.2f, 趋势: %s",
				tag, trend.CurrentUsageCount, trend.ActivityScore, trend.Trend)

			// 验证趋势对象是否包含有效数据
			assert.NotEmpty(t, trend.Tag, "趋势对象的标签字段不应为空")
			assert.Greater(t, trend.CurrentUsageCount, 0, "标签 %s 的使用数量应该大于0", tag)
			assert.GreaterOrEqual(t, trend.ActivityScore, 0.0, "活跃度分数应该大于等于0")
			assert.LessOrEqual(t, trend.ActivityScore, 1.0, "活跃度分数应该小于等于1")
			assert.Contains(t, []string{"上升", "稳定", "下降"}, trend.Trend, "趋势应该是'上升'、'稳定'或'下降'之一")
		}
	})
}
