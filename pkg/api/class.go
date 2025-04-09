package api

import (
	"context"
	"errors"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

func (c *Client) SearchByClassName(ctx context.Context, class string, limit int) ([]*response.Version, error) {
	if limit <= 0 {
		return c.IteratorByClassName(ctx, class).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetClassName(class)).SetLimit(limit)
		result, err := SearchRequestJsonDoc[*response.Version](c, ctx, search)
		if err != nil {
			return nil, err
		}
		if result == nil || result.ResponseBody == nil {
			return nil, errors.New("empty response body")
		}
		return result.ResponseBody.Docs, nil
	}
}

func (c *Client) IteratorByClassName(ctx context.Context, class string) *SearchIterator[*response.Version] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetClassName(class))
	return NewSearchIterator[*response.Version](search).WithClient(c)
}
