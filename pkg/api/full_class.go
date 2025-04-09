package api

import (
	"context"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// 根据全路径类名

func SearchByFullyQualifiedClassName(ctx context.Context, fullyQualifiedClassName string, limit int) ([]*response.Version, error) {
	if limit <= 0 {
		return IteratorByFullyQualifiedClassName(ctx, fullyQualifiedClassName).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetFullyQualifiedClassName(fullyQualifiedClassName)).SetLimit(limit)
		result, err := SearchRequest[*response.Version](ctx, search)
		if err != nil {
			return nil, err
		}
		return result.ResponseBody.Docs, nil
	}
}

func IteratorByFullyQualifiedClassName(ctx context.Context, fullyQualifiedClassName string) *SearchIterator[*response.Version] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetFullyQualifiedClassName(fullyQualifiedClassName))
	return NewSearchIterator[*response.Version](search)
}
