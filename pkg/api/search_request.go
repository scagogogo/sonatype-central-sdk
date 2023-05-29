package api

import (
	"context"
	"github.com/crawler-go-go-go/go-requests"
	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchRequest 底层API，构造查询参数进行列表查询
func SearchRequest[Doc any](ctx context.Context, searchRequest *request.SearchRequest) (*response.Response[Doc], error) {
	targetUrl := "https://search.maven.org/solrsearch/select?" + searchRequest.ToRequestParams()
	return requests.GetJson[*response.Response[Doc]](ctx, targetUrl)
}
