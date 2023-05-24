package api

import (
	"context"
	"github.com/scagogogo/sonatype-central-crawler/pkg/request"
	"github.com/scagogogo/sonatype-central-crawler/pkg/response"
)

func SearchByTag(ctx context.Context, tag string, limit int) ([]*response.Artifact, error) {
	if limit <= 0 {
		return IteratorByTag(ctx, tag).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetTags(tag)).SetLimit(limit)
		result, err := SearchRequest[*response.Artifact](ctx, search)
		if err != nil {
			return nil, err
		}
		return result.ResponseBody.Docs, nil
	}
}

func IteratorByTag(ctx context.Context, tag string) *SearchIterator[*response.Artifact] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetTags(tag))
	return NewSearchIterator[*response.Artifact](search)
}
