package api

import (
	"context"
	"fmt"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
)

// SearchRequest 执行搜索请求
// 基于SearchRequestJsonDoc实现
func (c *Client) SearchRequest(ctx context.Context, searchRequest *request.SearchRequest, result interface{}) error {
	targetUrl := fmt.Sprintf("%s/solrsearch/select?%s", c.baseURL, searchRequest.ToRequestParams())

	_, err := c.doRequest(ctx, "GET", targetUrl, nil, result)
	return err
}
