package api

import (
	"context"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// ListGAVs 列出符合条件的GAV（GroupId、ArtifactId、Version）
func (c *Client) ListGAVs(ctx context.Context, query string, limit int) ([]*response.Artifact, error) {
	searchRequest := request.NewSearchRequest()
	searchRequest.Query.SetCustomQuery(query)
	searchRequest.SetCore("gav")
	searchRequest.SetLimit(limit)
	searchRequest.AddCustomParam("wt", "json")

	var result response.Response[*response.Artifact]
	err := c.SearchRequest(ctx, searchRequest, &result)
	if err != nil {
		return nil, err
	}

	return result.ResponseBody.Docs, nil
}
