package api

import (
	"context"
	"github.com/scagogogo/sonatype-central-crawler/pkg/request"
	"github.com/scagogogo/sonatype-central-crawler/pkg/response"
)

// SearchByArtifactId 根据GroupID列出这个组下面的artifact
func SearchByArtifactId(ctx context.Context, artifactId string, limit int) ([]*response.Artifact, error) {
	if limit <= 0 {
		return IteratorByArtifactId(ctx, artifactId).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetArtifactId(artifactId)).SetLimit(limit)
		result, err := SearchRequest[*response.Artifact](ctx, search)
		if err != nil {
			return nil, err
		}
		return result.ResponseBody.Docs, nil
	}
}

func IteratorByArtifactId(ctx context.Context, artifactId string) *SearchIterator[*response.Artifact] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetArtifactId(artifactId))
	return NewSearchIterator[*response.Artifact](search)
}
