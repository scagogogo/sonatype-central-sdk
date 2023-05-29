package api

import (
	"context"
	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

func SearchBySha1(ctx context.Context, sha1 string, limit int) ([]*response.Version, error) {
	if limit <= 0 {
		return IteratorBySha1(ctx, sha1).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetSha1(sha1)).SetLimit(limit)
		result, err := SearchRequest[*response.Version](ctx, search)
		if err != nil {
			return nil, err
		}
		return result.ResponseBody.Docs, nil
	}
}

func IteratorBySha1(ctx context.Context, sha1 string) *SearchIterator[*response.Version] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetSha1(sha1))
	return NewSearchIterator[*response.Version](search)
}
