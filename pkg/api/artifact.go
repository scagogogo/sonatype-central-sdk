package api

import (
	"context"
	"errors"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchByArtifactId 根据ArtifactId列出这个组下面的artifact
func (c *Client) SearchByArtifactId(ctx context.Context, artifactId string, limit int) ([]*response.Artifact, error) {
	if limit <= 0 {
		return c.IteratorByArtifactId(ctx, artifactId).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetArtifactId(artifactId)).SetLimit(limit)
		result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, search)
		if err != nil {
			return nil, err
		}
		if result == nil || result.ResponseBody == nil {
			return nil, errors.New("empty response body")
		}
		return result.ResponseBody.Docs, nil
	}
}

func (c *Client) IteratorByArtifactId(ctx context.Context, artifactId string) *SearchIterator[*response.Artifact] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetArtifactId(artifactId))
	return NewSearchIterator[*response.Artifact](search).WithClient(c)
}
