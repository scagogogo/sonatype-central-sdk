package api

import (
	"context"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// AdvancedSearch 高级搜索，支持完整的坐标搜索
func (c *Client) AdvancedSearch(ctx context.Context, options *request.AdvancedSearchOptions, limit int) ([]*response.Artifact, error) {
	query := request.NewQuery()

	// 设置搜索参数
	if options.GroupId != "" {
		query.SetGroupId(options.GroupId)
	}

	if options.ArtifactId != "" {
		query.SetArtifactId(options.ArtifactId)
	}

	if options.Version != "" {
		query.SetVersion(options.Version)
	}

	if options.Packaging != "" {
		query.SetPackaging(options.Packaging)
	}

	if options.Classifier != "" {
		query.SetClassifier(options.Classifier)
	}

	// 创建搜索请求
	searchReq := request.NewSearchRequest().SetQuery(query).SetLimit(limit)

	// 执行搜索
	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, searchReq)
	if err != nil {
		return nil, err
	}

	return result.ResponseBody.Docs, nil
}

// AdvancedSearchIterator 高级搜索迭代器
func (c *Client) AdvancedSearchIterator(ctx context.Context, options *request.AdvancedSearchOptions) *SearchIterator[*response.Artifact] {
	query := request.NewQuery()

	// 设置搜索参数
	if options.GroupId != "" {
		query.SetGroupId(options.GroupId)
	}

	if options.ArtifactId != "" {
		query.SetArtifactId(options.ArtifactId)
	}

	if options.Version != "" {
		query.SetVersion(options.Version)
	}

	if options.Packaging != "" {
		query.SetPackaging(options.Packaging)
	}

	if options.Classifier != "" {
		query.SetClassifier(options.Classifier)
	}

	// 创建搜索请求
	searchReq := request.NewSearchRequest().SetQuery(query)

	return NewSearchIterator[*response.Artifact](searchReq)
}

// SearchWithSort 带排序的搜索函数
func (c *Client) SearchWithSort(ctx context.Context, searchQuery *request.SearchRequest, sortField string, ascending bool, limit int) ([]*response.Artifact, error) {
	// 设置排序
	searchQuery.SetSort(sortField, ascending)
	searchQuery.SetLimit(limit)

	// 执行搜索
	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, searchQuery)
	if err != nil {
		return nil, err
	}

	return result.ResponseBody.Docs, nil
}

// SearchByDependency 根据依赖搜索
func (c *Client) SearchByDependency(ctx context.Context, groupId, artifactId string, limit int) ([]*response.Artifact, error) {
	// 使用特殊查询格式搜索依赖
	query := request.NewQuery().
		SetCustomQuery(request.MakeDependencyQuery(groupId, artifactId))

	searchReq := request.NewSearchRequest().SetQuery(query).SetLimit(limit)

	// 执行搜索
	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, searchReq)
	if err != nil {
		return nil, err
	}

	return result.ResponseBody.Docs, nil
}

// SearchByLicense 根据许可证搜索
func (c *Client) SearchByLicense(ctx context.Context, license string, limit int) ([]*response.Artifact, error) {
	// 使用特殊查询格式搜索许可证
	query := request.NewQuery().
		SetCustomQuery(request.MakeLicenseQuery(license))

	searchReq := request.NewSearchRequest().SetQuery(query).SetLimit(limit)

	// 执行搜索
	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, searchReq)
	if err != nil {
		return nil, err
	}

	return result.ResponseBody.Docs, nil
}

// GetArtifactMetadata 获取完整的制品元数据
func (c *Client) GetArtifactMetadata(ctx context.Context, groupId, artifactId, version string) (*response.ArtifactMetadata, error) {
	// 使用GAV坐标查询
	query := request.NewQuery().
		SetGroupId(groupId).
		SetArtifactId(artifactId)

	if version != "" {
		query.SetVersion(version)
	}

	searchReq := request.NewSearchRequest().SetQuery(query).SetLimit(1)

	// 获取基本信息
	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, searchReq)
	if err != nil {
		return nil, err
	}

	if len(result.ResponseBody.Docs) == 0 {
		return nil, ErrNotFound
	}

	artifact := result.ResponseBody.Docs[0]

	// 如果有版本，尝试下载POM获取更详细信息
	metadata := &response.ArtifactMetadata{
		GroupId:       artifact.GroupId,
		ArtifactId:    artifact.ArtifactId,
		LatestVersion: artifact.LatestVersion,
		Packaging:     artifact.Packaging,
		LastUpdated:   artifact.Timestamp,
	}

	if version != "" {
		// 下载POM文件
		pomData, err := c.DownloadPom(ctx, groupId, artifactId, version)
		if err == nil {
			// 解析POM文件
			metadata.PomContent = string(pomData)
			// TODO: 解析POM获取更多元数据
		}
	}

	return metadata, nil
}

// BatchSearch 批量搜索多个制品
func (c *Client) BatchSearch(ctx context.Context, queries []*request.SearchRequest) (map[string][]*response.Artifact, error) {
	results := make(map[string][]*response.Artifact)

	// 创建结果通道
	type resultItem struct {
		key     string
		results []*response.Artifact
		err     error
	}

	resultChan := make(chan resultItem, len(queries))

	// 并发执行所有查询
	for i, query := range queries {
		go func(idx int, q *request.SearchRequest) {
			key := q.GetQueryKey()
			if key == "" {
				key = q.Query.ToRequestParamValue()
			}

			result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, q)
			if err != nil {
				resultChan <- resultItem{key: key, err: err}
				return
			}

			resultChan <- resultItem{key: key, results: result.ResponseBody.Docs}
		}(i, query)
	}

	// 收集结果
	for i := 0; i < len(queries); i++ {
		res := <-resultChan
		if res.err != nil {
			continue // 跳过错误的查询
		}
		results[res.key] = res.results
	}

	return results, nil
}
