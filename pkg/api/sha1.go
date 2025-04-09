package api

import (
	"context"
	"errors"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

func (c *Client) SearchBySha1(ctx context.Context, sha1 string, limit int) ([]*response.Version, error) {
	if limit <= 0 {
		return c.IteratorBySha1(ctx, sha1).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetSha1(sha1)).SetLimit(limit)
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

func (c *Client) IteratorBySha1(ctx context.Context, sha1 string) *SearchIterator[*response.Version] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetSha1(sha1))
	return NewSearchIterator[*response.Version](search).WithClient(c)
}
