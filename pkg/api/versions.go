package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

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

// GetLatestVersion 获取最新的发布版本
func (c *Client) GetLatestVersion(ctx context.Context, groupId, artifactId string) (*response.Version, error) {
	versions, err := c.ListVersions(ctx, groupId, artifactId, 1)
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for %s:%s", groupId, artifactId)
	}
	return versions[0], nil
}

// GetVersionsWithMetadata 获取所有版本并附带元数据信息
func (c *Client) GetVersionsWithMetadata(ctx context.Context, groupId, artifactId string) ([]*response.VersionWithMetadata, error) {
	versions, err := c.ListVersions(ctx, groupId, artifactId, 0)
	if err != nil {
		return nil, err
	}

	result := make([]*response.VersionWithMetadata, 0, len(versions))
	for _, version := range versions {
		versionInfo, err := c.GetVersionInfo(ctx, groupId, artifactId, version.Version)
		if err != nil {
			continue
		}

		result = append(result, &response.VersionWithMetadata{
			Version:     version,
			VersionInfo: versionInfo,
		})
	}

	return result, nil
}

// FilterVersions 根据条件过滤版本
func (c *Client) FilterVersions(ctx context.Context, groupId, artifactId string, filter func(*response.Version) bool) ([]*response.Version, error) {
	versions, err := c.ListVersions(ctx, groupId, artifactId, 0)
	if err != nil {
		return nil, err
	}

	result := make([]*response.Version, 0)
	for _, version := range versions {
		if filter(version) {
			result = append(result, version)
		}
	}

	return result, nil
}

// CompareVersions 比较两个版本
func (c *Client) CompareVersions(ctx context.Context, groupId, artifactId string, version1, version2 string) (*response.VersionComparison, error) {
	v1Info, err := c.GetVersionInfo(ctx, groupId, artifactId, version1)
	if err != nil {
		return nil, err
	}

	v2Info, err := c.GetVersionInfo(ctx, groupId, artifactId, version2)
	if err != nil {
		return nil, err
	}

	return &response.VersionComparison{
		Version1:    version1,
		Version2:    version2,
		V1Timestamp: v1Info.LastUpdated,
		V2Timestamp: v2Info.LastUpdated,
	}, nil
}

// HasVersion 检查特定版本是否存在
func (c *Client) HasVersion(ctx context.Context, groupId, artifactId, version string) (bool, error) {
	_, err := c.GetVersionInfo(ctx, groupId, artifactId, version)
	if err != nil {
		// 简单检查错误消息是否包含404或Not Found
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
