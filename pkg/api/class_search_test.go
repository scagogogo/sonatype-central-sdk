package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSearchClassesWithHighlighting 测试带高亮的类搜索功能
func TestSearchClassesWithHighlighting(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置更长的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 测试全限定类名
	className := "org.apache.commons.io.FileUtils"

	// 设置子测试的超时时间
	subCtx, subCancel := context.WithTimeout(ctx, 20*time.Second)
	defer subCancel()

	// 执行带高亮的搜索
	result, err := client.SearchClassesWithHighlighting(subCtx, className, 3)

	// 如果API连接失败，跳过测试
	if err != nil {
		t.Logf("搜索 %s 时出错: %v", className, err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	// 验证返回结果
	assert.NotNil(t, result, "搜索结果不应为空")
	assert.NotNil(t, result.ResponseBody, "响应体不应为空")
	assert.NotEmpty(t, result.ResponseBody.Docs, "搜索结果不应为空")

	// 验证高亮信息
	assert.NotNil(t, result.Highlighting, "高亮信息不应为空")

	// 输出结果信息
	t.Logf("找到 %d 个包含 %s 的结果", result.ResponseBody.NumFound, className)
	for i, doc := range result.ResponseBody.Docs {
		t.Logf("结果 %d: %s:%s:%s", i+1, doc.GroupId, doc.ArtifactId, doc.Version)

		// 显示高亮信息
		if highlightInfo, exists := result.Highlighting[doc.ID]; exists {
			if fchHighlights, hasFch := highlightInfo["fch"]; hasFch && len(fchHighlights) > 0 {
				t.Logf("  高亮: %s", fchHighlights[0])
			}
		}
	}
}

// TestExtractHighlightedClasses 测试高亮提取函数
func TestExtractHighlightedClasses(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 测试全限定类名
	className := "org.junit.Assert"

	// 设置子测试的超时时间
	subCtx, subCancel := context.WithTimeout(ctx, 20*time.Second)
	defer subCancel()

	// 执行带高亮的搜索
	result, err := client.SearchClassesWithHighlighting(subCtx, className, 3)

	// 如果API连接失败，跳过测试
	if err != nil {
		t.Logf("搜索 %s 时出错: %v", className, err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	// 提取高亮信息
	highlights := ExtractHighlightedClasses(result)

	// 验证提取的高亮信息
	assert.NotNil(t, highlights, "提取的高亮信息不应为空")
	assert.NotEmpty(t, highlights, "提取的高亮信息应该包含数据")

	// 输出结果信息
	t.Logf("从 %d 个结果中提取了 %d 个高亮类名", len(result.ResponseBody.Docs), len(highlights))
	for docId, highlightedClasses := range highlights {
		t.Logf("文档 ID: %s", docId)
		for i, cls := range highlightedClasses {
			t.Logf("  高亮类 %d: %s", i+1, cls)
		}
	}
}

// TestSearchFullyQualifiedClassNames 测试完全限定类名搜索
func TestSearchFullyQualifiedClassNames(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 测试全限定类名
	className := "org.apache.commons.lang3.StringUtils"

	// 设置子测试的超时时间
	subCtx, subCancel := context.WithTimeout(ctx, 20*time.Second)
	defer subCancel()

	// 执行搜索
	versions, highlights, err := client.SearchFullyQualifiedClassNames(subCtx, className, 3)

	// 如果API连接失败，跳过测试
	if err != nil {
		t.Logf("搜索 %s 时出错: %v", className, err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	// 验证搜索结果
	assert.NotNil(t, versions, "搜索结果不应为空")
	assert.NotEmpty(t, versions, "搜索结果应该包含数据")
	assert.NotNil(t, highlights, "高亮信息不应为空")

	// 输出结果信息
	t.Logf("找到 %d 个包含 %s 的结果", len(versions), className)
	for i, v := range versions {
		t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)

		// 显示对应的高亮信息
		if hlClasses, exists := highlights[v.ID]; exists {
			for j, cls := range hlClasses {
				t.Logf("  高亮类 %d: %s", j+1, cls)
			}
		}
	}
}
