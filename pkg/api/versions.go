package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// GetVersionInfo 获取特定版本的详细信息
func (c *Client) GetVersionInfo(ctx context.Context, groupId, artifactId, version string) (*response.VersionInfo, error) {
	url := fmt.Sprintf("%s/v1/versions/%s/%s/%s", c.GetBaseURL(), groupId, artifactId, version)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var versionInfo response.VersionInfo
	if err := c.doRequest(req, &versionInfo); err != nil {
		return nil, err
	}

	return &versionInfo, nil
}

// ListVersions 根据GroupID和artifactId列出下面的所有版本
func (c *Client) ListVersions(ctx context.Context, groupId, artifactId string, limit int) ([]*response.Version, error) {
	if limit <= 0 {
		return c.IteratorVersions(ctx, groupId, artifactId).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId(groupId).SetArtifactId(artifactId)).SetCore("gav").SetLimit(limit)
		result, err := SearchRequestJsonDoc[*response.Version](c, ctx, search)
		if err != nil {
			return nil, err
		}
		if result == nil || result.ResponseBody == nil {
			return nil, errors.New("empty response body")
		}
		return result.ResponseBody.Docs, nil
	}
}

// IteratorVersions 返回一个版本迭代器
func (c *Client) IteratorVersions(ctx context.Context, groupId, artifactId string) *SearchIterator[*response.Version] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId(groupId).SetArtifactId(artifactId)).SetCore("gav")
	return NewSearchIterator[*response.Version](search).WithClient(c)
}
