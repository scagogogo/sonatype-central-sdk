package api

import (
	"context"
	"fmt"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchRequestJsonDoc 执行搜索请求并解析JSON响应
// 这是一个通用函数，处理所有类型的文档搜索请求
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
