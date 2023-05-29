package api

import (
	"context"
	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchByGroupId 根据GroupID列出这个组下面的artifact
func SearchByGroupId(ctx context.Context, groupId string, limit int) ([]*response.Artifact, error) {
	if limit <= 0 {
		return IteratorByGroupId(ctx, groupId).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId(groupId)).SetLimit(limit)
		result, err := SearchRequest[*response.Artifact](ctx, search)
		if err != nil {
			return nil, err
		}
		return result.ResponseBody.Docs, nil
	}
}

func IteratorByGroupId(ctx context.Context, groupId string) *SearchIterator[*response.Artifact] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId(groupId))
	return NewSearchIterator[*response.Artifact](search)
}
