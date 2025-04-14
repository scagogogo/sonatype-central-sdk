package api

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
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

	// 验证返回结果的基本结构
	assert.NotNil(t, result, "搜索结果不应为空")
	assert.NotNil(t, result.ResponseHeader, "响应头不应为空")
	assert.NotNil(t, result.ResponseBody, "响应体不应为空")
	assert.NotEmpty(t, result.ResponseBody.Docs, "搜索结果不应为空")

	// 验证响应状态
	assert.Equal(t, 0, result.ResponseHeader.Status, "响应状态应为成功(0)")
	assert.Greater(t, result.ResponseBody.NumFound, 0, "应找到至少一个结果")

	// 验证高亮信息
	assert.NotNil(t, result.Highlighting, "高亮信息不应为空")
	assert.GreaterOrEqual(t, len(result.Highlighting), 1, "至少应有一个文档包含高亮信息")

	// 验证返回的文档和高亮的匹配关系
	for _, doc := range result.ResponseBody.Docs {
		// 检查文档ID是否在高亮映射中
		if highlightInfo, exists := result.Highlighting[doc.ID]; exists {
			// 检查fch字段是否存在
			if fchHighlights, hasFch := highlightInfo["fch"]; hasFch {
				assert.NotEmpty(t, fchHighlights, "高亮列表不应为空")

				// 验证高亮文本是否包含<em>标签
				for _, highlight := range fchHighlights {
					assert.Contains(t, highlight, "<em>", "高亮文本应包含<em>标签")
					assert.Contains(t, highlight, "</em>", "高亮文本应包含</em>标签")

					// 验证高亮文本包含搜索词(不区分大小写)
					searchTermLower := strings.ToLower(strings.Replace(className, ".", "\\.", -1))
					assert.Contains(t, strings.ToLower(highlight), searchTermLower,
						"高亮文本应包含搜索词")
				}
			}
		}
	}

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

	// 验证文档数量匹配
	docCount := 0
	for _, doc := range result.ResponseBody.Docs {
		if _, exists := result.Highlighting[doc.ID]; exists {
			docCount++
		}
	}

	assert.Equal(t, docCount, len(highlights), "提取的高亮条目数应与原始响应中的文档数匹配")

	// 验证每个文档的提取结果
	for docId, classHighlights := range highlights {
		// 原始高亮信息
		origHighlights := result.Highlighting[docId]["fch"]

		// 验证提取后的结果与原始结果一致
		assert.Equal(t, len(origHighlights), len(classHighlights),
			"提取后的高亮数量应与原始高亮数量一致")

		for i, hl := range origHighlights {
			assert.Equal(t, hl, classHighlights[i], "提取的高亮内容应与原始高亮内容一致")
		}
	}

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
	assert.NotEmpty(t, highlights, "高亮信息应包含数据")

	// 验证文档和高亮信息的对应关系
	for _, version := range versions {
		// 检查每个版本是否都有对应的高亮信息
		hlClasses, exists := highlights[version.ID]
		assert.True(t, exists, "每个版本都应该有对应的高亮信息")
		assert.NotEmpty(t, hlClasses, "高亮类名列表不应为空")

		// 验证高亮内容
		for _, hlClass := range hlClasses {
			assert.Contains(t, hlClass, "<em>", "高亮文本应包含<em>标签")
			assert.Contains(t, hlClass, "</em>", "高亮文本应包含</em>标签")
		}
	}

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

// TestHighlightingEdgeCases 测试高亮搜索的边界情况
func TestHighlightingEdgeCases(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置基本超时
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 定义测试用例
	testCases := []struct {
		name      string
		className string
		limit     int
		expectErr bool
	}{
		{"空类名", "", 3, false},                      // 空类名不一定会错误，可能返回默认结果
		{"特殊字符", "org.apache.commons.*", 3, false}, // 星号是有效的通配符
		{"非常小的限制", "java.lang.String", 1, false},   // 限制为1应该正常工作
		{"零限制", "java.util.List", 0, false},        // 零限制应使用默认值
		{"负限制", "org.junit.Test", -1, false},       // 负限制应使用默认值
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 设置子测试超时
			subCtx, subCancel := context.WithTimeout(ctx, 10*time.Second)
			defer subCancel()

			// 执行高亮搜索
			result, err := client.SearchClassesWithHighlighting(subCtx, tc.className, tc.limit)

			// 检查错误
			if tc.expectErr {
				assert.Error(t, err, "期望出现错误")
			} else if err != nil {
				t.Logf("搜索 %s 时出错: %v", tc.className, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			// 验证返回结果
			assert.NotNil(t, result, "即使是边界情况，结果也不应为空")
			assert.NotNil(t, result.ResponseHeader, "响应头不应为空")
			assert.NotNil(t, result.ResponseBody, "响应体不应为空")

			// 对于零/负限制的情况，验证是否使用了默认限制
			if tc.limit <= 0 && len(result.ResponseBody.Docs) > 0 {
				t.Logf("使用默认限制时返回了 %d 个结果", len(result.ResponseBody.Docs))
			}

			// 检查限制为1的情况
			if tc.limit == 1 && len(result.ResponseBody.Docs) > 0 {
				assert.Equal(t, 1, len(result.ResponseBody.Docs), "当限制为1时，应只返回1个结果")
			}

			t.Logf("边界情况 %s: 找到 %d 个结果", tc.name, result.ResponseBody.NumFound)
		})
	}
}

// TestExtractHighlightedClassesEdgeCases 测试高亮提取函数的边界情况
func TestExtractHighlightedClassesEdgeCases(t *testing.T) {
	// 测试空输入
	t.Run("空输入", func(t *testing.T) {
		result := ExtractHighlightedClasses(nil)
		assert.Nil(t, result, "对于空输入，应返回nil")
	})

	// 测试没有高亮信息的Response
	t.Run("无高亮信息", func(t *testing.T) {
		response := &response.Response[*response.Version]{
			ResponseHeader: &response.ResponseHeader{},
			ResponseBody:   &response.ResponseBody[*response.Version]{},
			Highlighting:   nil,
		}

		result := ExtractHighlightedClasses(response)
		assert.Nil(t, result, "对于没有高亮信息的响应，应返回nil")
	})

	// 测试空高亮映射
	t.Run("空高亮映射", func(t *testing.T) {
		response := &response.Response[*response.Version]{
			ResponseHeader: &response.ResponseHeader{},
			ResponseBody:   &response.ResponseBody[*response.Version]{},
			Highlighting:   make(map[string]map[string][]string),
		}

		result := ExtractHighlightedClasses(response)
		assert.NotNil(t, result, "对于空高亮映射，应返回非nil的映射")
		assert.Empty(t, result, "对于空高亮映射，返回的映射应为空")
	})
}
