package api

import (
	"context"
	"errors"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchClassesWithHighlighting 根据完全限定类名搜索制品，并返回高亮信息
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - fullyQualifiedClassName: 完全限定类名，如"org.specs.runner.JUnit"
//   - limit: 最大返回结果数量，如果小于等于0则不限制
//
// 返回:
//   - 搜索结果: 包含匹配的版本信息以及高亮片段
//   - 错误: 如果搜索过程中发生错误
func (c *Client) SearchClassesWithHighlighting(ctx context.Context, fullyQualifiedClassName string, limit int) (*response.Response[*response.Version], error) {
	// 创建搜索请求
	query := request.NewQuery().SetFullyQualifiedClassName(fullyQualifiedClassName)
	searchReq := request.NewSearchRequest().SetQuery(query)

	// 设置限制条数
	if limit > 0 {
		searchReq.SetLimit(limit)
	}

	// 添加高亮相关参数
	searchReq.AddCustomParam("hl", "true")
	searchReq.AddCustomParam("hl.fl", "fch")     // 高亮完全限定类名字段
	searchReq.AddCustomParam("hl.snippets", "3") // 最多返回3个高亮片段

	// 执行搜索请求
	result, err := SearchRequestJsonDoc[*response.Version](c, ctx, searchReq)
	if err != nil {
		return nil, err
	}

	// 验证响应
	if result == nil || result.ResponseBody == nil {
		return nil, errors.New("empty response body")
	}

	return result, nil
}

// ExtractHighlightedClasses 从搜索结果中提取出每个制品中的高亮类名
// 参数:
//   - result: 带有高亮信息的搜索结果
//
// 返回:
//   - 映射关系: 文档ID到该文档中包含的高亮类名列表
func ExtractHighlightedClasses(result *response.Response[*response.Version]) map[string][]string {
	if result == nil || result.Highlighting == nil {
		return nil
	}

	// 创建结果映射
	highlightedClasses := make(map[string][]string)

	// 遍历每个文档的高亮信息
	for docId, fields := range result.Highlighting {
		// 提取类字段(fch)的高亮片段
		if classHighlights, exists := fields["fch"]; exists && len(classHighlights) > 0 {
			highlightedClasses[docId] = classHighlights
		}
	}

	return highlightedClasses
}

// SearchFullyQualifiedClassNames 搜索完全限定类名并返回包含这些类的所有制品版本
// 参数:
//   - ctx: 上下文，可用于取消或设置超时
//   - className: 要搜索的完全限定类名或部分类名（支持模式匹配）
//   - limit: 最大返回结果数量
//
// 返回:
//   - 版本列表: 包含匹配类的所有制品版本
//   - 高亮信息: 每个制品中匹配的具体类名
//   - 错误: 如果搜索过程中发生错误
func (c *Client) SearchFullyQualifiedClassNames(ctx context.Context, className string, limit int) ([]*response.Version, map[string][]string, error) {
	// 执行带高亮的搜索
	result, err := c.SearchClassesWithHighlighting(ctx, className, limit)
	if err != nil {
		return nil, nil, err
	}

	// 提取高亮的类名
	highlights := ExtractHighlightedClasses(result)

	return result.ResponseBody.Docs, highlights, nil
}
