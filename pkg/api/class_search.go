package api

import (
	"context"
	"errors"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchClassesWithHighlighting 根据完全限定类名搜索制品，并返回高亮信息
//
// 该方法在Maven Central仓库中搜索包含指定类的所有制品，并在结果中高亮显示匹配的类名。
// 高亮信息存储在Response.Highlighting字段中，格式为map[文档ID]map[字段名][]高亮片段。
// 通常，"fch"字段包含全限定类名的高亮信息，其中匹配部分会用<em>标签包围。
//
// 参数:
//   - ctx: 上下文，可用于取消或设置超时，建议设置30秒以上的超时
//   - fullyQualifiedClassName: 完全限定类名，如"org.specs.runner.JUnit"，支持部分匹配
//   - limit: 最大返回结果数量，推荐1-50，如果<=0则使用服务器默认值(10)
//
// 返回:
//   - 搜索结果: 包含匹配的版本信息(ResponseBody.Docs)以及高亮片段(Highlighting)
//   - 错误: 连接错误、超时或服务端错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	result, err := client.SearchClassesWithHighlighting(ctx, "org.apache.commons.io.FileUtils", 5)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 处理高亮信息
//	for docId, fields := range result.Highlighting {
//	    for _, highlight := range fields["fch"] {
//	        fmt.Printf("文档 %s 包含类: %s\n", docId, highlight)
//	    }
//	}
func (c *Client) SearchClassesWithHighlighting(ctx context.Context, fullyQualifiedClassName string, limit int) (*response.Response[*response.Version], error) {
	// 步骤1: 创建搜索请求对象
	query := request.NewQuery().SetFullyQualifiedClassName(fullyQualifiedClassName)
	searchReq := request.NewSearchRequest().SetQuery(query)

	// 步骤2: 设置限制条数，如果用户不指定，使用服务器默认值
	if limit > 0 {
		searchReq.SetLimit(limit)
	}

	// 步骤3: 添加高亮相关参数
	// hl=true: 启用高亮功能
	searchReq.AddCustomParam("hl", "true")
	// hl.fl=fch: 指定在"fch"(fully qualified class name)字段上启用高亮
	searchReq.AddCustomParam("hl.fl", "fch")
	// hl.snippets=3: 每个匹配最多返回3个高亮片段
	searchReq.AddCustomParam("hl.snippets", "3")

	// 步骤4: 执行搜索请求
	result, err := SearchRequestJsonDoc[*response.Version](c, ctx, searchReq)
	if err != nil {
		return nil, err
	}

	// 步骤5: 验证响应是否有效
	if result == nil || result.ResponseBody == nil {
		return nil, errors.New("empty response body")
	}

	return result, nil
}

// ExtractHighlightedClasses 从搜索结果中提取出每个制品中的高亮类名
//
// 该方法将Response.Highlighting字段中的高亮信息提取为更简单的映射结构，
// 便于在应用中使用。结果为每个文档ID到其匹配的高亮类名列表的映射。
//
// 参数:
//   - result: 带有高亮信息的搜索结果，通常由SearchClassesWithHighlighting方法返回
//
// 返回:
//   - 映射关系: map[文档ID][]高亮类名
//     例如：{"org.apache:commons-io:1.2.3": ["<em>org.apache</em>.commons.io.FileUtils"]}
//
// 使用示例:
//
//	result, _ := client.SearchClassesWithHighlighting(ctx, "org.apache.commons.io", 5)
//	highlights := api.ExtractHighlightedClasses(result)
//
//	for docId, classNames := range highlights {
//	    fmt.Printf("文档 %s 包含以下类:\n", docId)
//	    for _, className := range classNames {
//	        fmt.Printf("  - %s\n", className)
//	    }
//	}
func ExtractHighlightedClasses(result *response.Response[*response.Version]) map[string][]string {
	// 步骤1: 验证输入是否有效
	if result == nil || result.Highlighting == nil {
		return nil
	}

	// 步骤2: 创建结果映射
	highlightedClasses := make(map[string][]string)

	// 步骤3: 遍历每个文档的高亮信息
	for docId, fields := range result.Highlighting {
		// 步骤4: 提取类字段(fch)的高亮片段
		// 检查fch字段是否存在且有值
		if classHighlights, exists := fields["fch"]; exists && len(classHighlights) > 0 {
			highlightedClasses[docId] = classHighlights
		}
	}

	return highlightedClasses
}

// SearchFullyQualifiedClassNames 搜索完全限定类名并返回包含这些类的所有制品版本
//
// 该方法是SearchClassesWithHighlighting的便捷包装，提供更直接的接口返回版本列表和高亮信息。
// 特别适用于需要同时获取版本详情和匹配类名的场景。
//
// 参数:
//   - ctx: 上下文，可用于取消或设置超时，建议设置30秒以上的超时
//   - className: 要搜索的完全限定类名或部分类名，如"org.apache.commons.io"
//   - limit: 最大返回结果数量，推荐1-50，如果<=0则使用服务器默认值
//
// 返回:
//   - 版本列表: 包含匹配类的所有制品版本信息，每个版本包含groupId、artifactId、version等
//   - 高亮信息: 每个制品中匹配的具体类名，格式为map[文档ID][]高亮类名
//   - 错误: 连接错误、超时或服务端错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	versions, highlights, err := client.SearchFullyQualifiedClassNames(ctx, "org.junit.Assert", 10)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	for i, version := range versions {
//	    fmt.Printf("%d. %s:%s:%s\n", i+1, version.GroupId, version.ArtifactId, version.Version)
//
//	    // 显示每个版本匹配的类
//	    if classes, ok := highlights[version.ID]; ok {
//	        for _, cls := range classes {
//	            fmt.Printf("   - %s\n", cls)
//	        }
//	    }
//	}
func (c *Client) SearchFullyQualifiedClassNames(ctx context.Context, className string, limit int) ([]*response.Version, map[string][]string, error) {
	// 步骤1: 执行带高亮的搜索
	// 这里使用前面定义的SearchClassesWithHighlighting方法
	result, err := c.SearchClassesWithHighlighting(ctx, className, limit)
	if err != nil {
		return nil, nil, err
	}

	// 步骤2: 提取高亮的类名
	// 使用前面定义的ExtractHighlightedClasses方法
	highlights := ExtractHighlightedClasses(result)

	// 步骤3: 返回文档列表和高亮信息
	return result.ResponseBody.Docs, highlights, nil
}
