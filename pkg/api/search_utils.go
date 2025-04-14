package api

import (
	"context"
	"fmt"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchRequestJsonDoc 执行搜索请求并将JSON响应解析为指定类型的结构体
//
// 该函数是SDK中所有搜索操作的核心实现，它接收一个搜索请求，构建URL，执行HTTP GET请求，
// 并将返回的JSON响应解析为泛型类型Response[Doc]。函数利用Go泛型特性，可以适应不同的
// 文档类型（Artifact、Version等），使API调用更加类型安全。
//
// 参数:
//   - c: API客户端实例，包含基础URL和HTTP配置；如果为nil，将使用默认客户端
//   - ctx: 请求上下文，用于控制超时和取消
//   - searchRequest: 包含查询参数、分页、排序等信息的搜索请求对象
//
// 返回:
//   - *response.Response[Doc]: 解析后的响应结构，包含文档列表、分面统计等
//   - error: 如果请求失败、超时或解析出错时返回相应错误
//
// 使用示例:
//
//	// 搜索GroupId为"org.apache.commons"的所有制品
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 构建搜索请求
//	query := request.NewQuery().SetGroupId("org.apache.commons")
//	searchReq := request.NewSearchRequest().SetQuery(query).SetLimit(10)
//
//	// 执行搜索
//	result, err := api.SearchRequestJsonDoc[*response.Artifact](client, ctx, searchReq)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 处理搜索结果
//	fmt.Printf("找到 %d 个结果\n", result.ResponseBody.NumFound)
//	for _, artifact := range result.ResponseBody.Docs {
//	    fmt.Printf("%s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
func SearchRequestJsonDoc[Doc any](c *Client, ctx context.Context, searchRequest *request.SearchRequest) (*response.Response[Doc], error) {
	targetUrl := fmt.Sprintf("%s/solrsearch/select?%s", c.baseURL, searchRequest.ToRequestParams())

	var result response.Response[Doc]
	_, err := c.doRequest(ctx, "GET", targetUrl, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// SearchRequestJson 是SearchRequestJsonDoc的别名，保持向后兼容
// 请使用SearchRequestJsonDoc代替此函数
func SearchRequestJson[Doc any](c *Client, ctx context.Context, searchRequest *request.SearchRequest) (*response.Response[Doc], error) {
	return SearchRequestJsonDoc[Doc](c, ctx, searchRequest)
}
