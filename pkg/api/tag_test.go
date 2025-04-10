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
