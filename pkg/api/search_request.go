package api

import (
	"context"
	"fmt"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
)

// SearchRequest 执行搜索请求并将结果解析到指定的结构体中
//
// 该方法是SDK中搜索功能的底层实现之一，它接收一个搜索请求对象，构建完整的URL，
// 执行HTTP GET请求，并将返回的JSON响应解析到提供的result结构体中。与SearchRequestJsonDoc不同，
// 此方法需要调用者提供用于接收结果的结构体实例。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - searchRequest: 包含查询参数、分页、排序等信息的搜索请求对象
//   - result: 用于存储解析后响应的结构体指针，必须是一个有效的指针类型
//
// 返回:
//   - error: 如果请求失败或解析出错时返回相应错误；成功时返回nil
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 构建搜索请求
//	query := request.NewQuery().SetGroupId("org.apache.commons")
//	searchReq := request.NewSearchRequest().SetQuery(query).SetLimit(10)
//
//	// 准备接收结果的结构体
//	var result response.Response[*response.Artifact]
//
//	// 执行搜索
//	err := client.SearchRequest(ctx, searchReq, &result)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 处理搜索结果
//	fmt.Printf("找到 %d 个结果\n", result.ResponseBody.NumFound)
//	for _, artifact := range result.ResponseBody.Docs {
//	    fmt.Printf("%s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
func (c *Client) SearchRequest(ctx context.Context, searchRequest *request.SearchRequest, result interface{}) error {
	targetUrl := fmt.Sprintf("%s/solrsearch/select?%s", c.baseURL, searchRequest.ToRequestParams())

	_, err := c.doRequest(ctx, "GET", targetUrl, nil, result)
	return err
}
