package api

import (
	"context"
	"github.com/scagogogo/sonatype-central-crawler/pkg/request"
	"github.com/scagogogo/sonatype-central-crawler/pkg/response"
)

func SearchByClassName(ctx context.Context, class string, limit int) ([]*response.Version, error) {
	if limit <= 0 {
		return IteratorByClassName(ctx, class).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetClassName(class)).SetLimit(limit)
		result, err := SearchRequest[*response.Version](ctx, search)
		if err != nil {
			return nil, err
		}
		return result.ResponseBody.Docs, nil
	}
}

func IteratorByClassName(ctx context.Context, class string) *SearchIterator[*response.Version] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetClassName(class))
	return NewSearchIterator[*response.Version](search)
}
