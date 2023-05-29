package api

import (
	"context"
	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchByGAV 根据GroupID列出这个组下面的artifact
func SearchByGAV(ctx context.Context, groupId, artifactId, version string) ([]*response.Artifact, error) {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId(groupId).SetArtifactId(artifactId).SetVersion(version))
	result, err := SearchRequest[*response.Artifact](ctx, search)
	if err != nil {
		return nil, err
	}
	return result.ResponseBody.Docs, nil
}
