package api

import (
	"context"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// AsyncSearchByGroup 异步搜索组ID
func (c *Client) AsyncSearchByGroup(ctx context.Context, groupId string, limit int) <-chan AsyncResult[[]*response.Artifact] {
	resultChan := make(chan AsyncResult[[]*response.Artifact], 1)

	go func() {
		defer close(resultChan)

		result, err := c.SearchByGroupId(ctx, groupId, limit)
		resultChan <- AsyncResult[[]*response.Artifact]{
			Result:  result,
			Error:   err,
			Context: groupId,
		}
	}()

	return resultChan
}

// AsyncSearchByArtifact 异步搜索制品ID
func (c *Client) AsyncSearchByArtifact(ctx context.Context, artifactId string, limit int) <-chan AsyncResult[[]*response.Artifact] {
	resultChan := make(chan AsyncResult[[]*response.Artifact], 1)

	go func() {
		defer close(resultChan)

		result, err := c.SearchByArtifactId(ctx, artifactId, limit)
		resultChan <- AsyncResult[[]*response.Artifact]{
			Result:  result,
			Error:   err,
			Context: artifactId,
		}
	}()

	return resultChan
}

// AsyncBatchSearch 批量异步搜索
func (c *Client) AsyncBatchSearch(ctx context.Context, requests []*request.SearchRequest) <-chan AsyncResult[[]*response.Artifact] {
	resultChan := make(chan AsyncResult[[]*response.Artifact], len(requests))

	// 启动所有请求
	for _, req := range requests {
		go func(searchReq *request.SearchRequest) {
			result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, searchReq)
			if err != nil {
				resultChan <- AsyncResult[[]*response.Artifact]{
					Error:   err,
					Context: searchReq,
				}
				return
			}

			resultChan <- AsyncResult[[]*response.Artifact]{
				Result:  result.ResponseBody.Docs,
				Context: searchReq,
			}
		}(req)
	}

	return resultChan
}

// AsyncGetArtifactMetadata 异步获取制品元数据
func (c *Client) AsyncGetArtifactMetadata(ctx context.Context, groupId, artifactId, version string) <-chan AsyncResult[*response.ArtifactMetadata] {
	resultChan := make(chan AsyncResult[*response.ArtifactMetadata], 1)

	go func() {
		defer close(resultChan)

		result, err := c.GetArtifactMetadata(ctx, groupId, artifactId, version)
		resultChan <- AsyncResult[*response.ArtifactMetadata]{
			Result: result,
			Error:  err,
			Context: map[string]string{
				"groupId":    groupId,
				"artifactId": artifactId,
				"version":    version,
			},
		}
	}()

	return resultChan
}
