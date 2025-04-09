package api

import (
	"context"
	"errors"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchByTag 根据标签搜索项目
func (c *Client) SearchByTag(ctx context.Context, tag string, limit int) ([]*response.Artifact, error) {
	if limit <= 0 {
		return c.IteratorByTag(ctx, tag).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetTags(tag)).SetLimit(limit)
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

// IteratorByTag 返回根据标签搜索的迭代器
func (c *Client) IteratorByTag(ctx context.Context, tag string) *SearchIterator[*response.Artifact] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetTags(tag))
	return NewSearchIterator[*response.Artifact](search).WithClient(c)
}
