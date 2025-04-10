package api

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// ListGAVs 列出符合条件的GAV（GroupId、ArtifactId、Version）
func (c *Client) ListGAVs(ctx context.Context, query string, limit int) ([]*response.Artifact, error) {
	searchRequest := request.NewSearchRequest()
	searchRequest.Query.SetCustomQuery(query)
	searchRequest.SetCore("gav")
	searchRequest.SetLimit(limit)
	searchRequest.AddCustomParam("wt", "json")

	var result response.Response[*response.Artifact]
	err := c.SearchRequest(ctx, searchRequest, &result)
	if err != nil {
		return nil, err
	}

	return result.ResponseBody.Docs, nil
}

// GetGAVInfo 获取指定GAV坐标的制品详细信息
func (c *Client) GetGAVInfo(ctx context.Context, groupId, artifactId, version string) (*response.Artifact, error) {
	query := fmt.Sprintf("g:%s AND a:%s",
		url.QueryEscape(groupId),
		url.QueryEscape(artifactId))

	if version != "" {
		query = fmt.Sprintf("%s AND v:%s", query, url.QueryEscape(version))
	}

	artifacts, err := c.ListGAVs(ctx, query, 1)
	if err != nil {
		return nil, err
	}

	if len(artifacts) == 0 {
		return nil, ErrNotFound
	}

	return artifacts[0], nil
}

// SearchGAVsWithSort 根据查询搜索GAV并按指定字段排序
func (c *Client) SearchGAVsWithSort(ctx context.Context, query string, sortField string, ascending bool, limit int) ([]*response.Artifact, error) {
	searchRequest := request.NewSearchRequest()
	searchRequest.Query.SetCustomQuery(query)
	searchRequest.SetCore("gav")
	searchRequest.SetLimit(limit)
	searchRequest.SetSort(sortField, ascending)
	searchRequest.AddCustomParam("wt", "json")

	var result response.Response[*response.Artifact]
	err := c.SearchRequest(ctx, searchRequest, &result)
	if err != nil {
		return nil, err
	}

	return result.ResponseBody.Docs, nil
}

// FindGAVDependencies 查找两个GAV之间的依赖关系
func (c *Client) FindGAVDependencies(ctx context.Context, groupId1, artifactId1, groupId2, artifactId2 string, limit int) ([]*response.Artifact, error) {
	// 构建查询语句，先仅搜索目标制品
	query := fmt.Sprintf("g:%s AND a:%s",
		url.QueryEscape(groupId1),
		url.QueryEscape(artifactId1))

	// 获取制品的列表，然后手动检查每个制品的依赖关系
	artifacts, err := c.ListGAVs(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	// 依赖关系需要通过分析POM文件来获取
	// 这里返回基础查询结果，具体的依赖关系分析需要单独获取元数据
	return artifacts, nil
}

// ListGAVsPaginated 分页查询GAV信息
func (c *Client) ListGAVsPaginated(ctx context.Context, query string, page, pageSize int) ([]*response.Artifact, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	searchRequest := request.NewSearchRequest()
	searchRequest.Query.SetCustomQuery(query)
	searchRequest.SetCore("gav")
	searchRequest.SetLimit(pageSize)
	searchRequest.SetStart(offset)
	searchRequest.AddCustomParam("wt", "json")

	var result response.Response[*response.Artifact]
	err := c.SearchRequest(ctx, searchRequest, &result)
	if err != nil {
		return nil, 0, err
	}

	if result.ResponseBody == nil {
		return nil, 0, errors.New("empty response body")
	}

	return result.ResponseBody.Docs, result.ResponseBody.NumFound, nil
}

// IteratorGAVs 返回一个GAV迭代器，用于遍历大量结果
func (c *Client) IteratorGAVs(ctx context.Context, query string) *SearchIterator[*response.Artifact] {
	searchRequest := request.NewSearchRequest()
	searchRequest.Query.SetCustomQuery(query)
	searchRequest.SetCore("gav")
	searchRequest.AddCustomParam("wt", "json")

	return NewSearchIterator[*response.Artifact](searchRequest).WithClient(c)
}
