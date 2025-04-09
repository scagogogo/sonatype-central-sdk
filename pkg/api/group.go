package api

import (
	"context"
	"errors"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchByGroupId 根据GroupID列出这个组下面的artifact
func (c *Client) SearchByGroupId(ctx context.Context, groupId string, limit int) ([]*response.Artifact, error) {
	if limit <= 0 {
		return c.IteratorByGroupId(ctx, groupId).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId(groupId)).SetLimit(limit)
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

func (c *Client) IteratorByGroupId(ctx context.Context, groupId string) *SearchIterator[*response.Artifact] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId(groupId))
	return NewSearchIterator[*response.Artifact](search).WithClient(c)
}
