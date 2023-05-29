package api

import (
	"context"
	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// ListVersions 根据GroupID和artifactId列出下面的所有版本
func ListVersions(ctx context.Context, groupId, artifactId string, limit int) ([]*response.Version, error) {
	if limit <= 0 {
		return IteratorVersions(ctx, groupId, artifactId).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId(groupId).SetArtifactId(artifactId)).SetCore("gav").SetLimit(limit)
		result, err := SearchRequest[*response.Version](ctx, search)
		if err != nil {
			return nil, err
		}
		return result.ResponseBody.Docs, nil
	}
}

func IteratorVersions(ctx context.Context, groupId, artifactId string) *SearchIterator[*response.Version] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId(groupId).SetArtifactId(artifactId)).SetCore("gav")
	return NewSearchIterator[*response.Version](search)
}
