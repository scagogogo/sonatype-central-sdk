package api

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchByArtifactId 根据ArtifactId列出这个组下面的artifact
func (c *Client) SearchByArtifactId(ctx context.Context, artifactId string, limit int) ([]*response.Artifact, error) {
	if limit <= 0 {
		return c.IteratorByArtifactId(ctx, artifactId).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetArtifactId(artifactId)).SetLimit(limit)
		result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, search)
		if err != nil {
			return nil, err
		}
		if result == nil || result.ResponseBody == nil {
			return nil, errors.New("empty response body")
		}
		return result.ResponseBody.Docs, nil
	}
}

func (c *Client) IteratorByArtifactId(ctx context.Context, artifactId string) *SearchIterator[*response.Artifact] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetArtifactId(artifactId))
	return NewSearchIterator[*response.Artifact](search).WithClient(c)
}

// SearchByGroupAndArtifactId 根据组ID和制品ID精确搜索制品
// 参数:
//   - ctx: 上下文
//   - groupId: 组ID
//   - artifactId: 制品ID
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - 制品列表
//   - 错误信息
func (c *Client) SearchByGroupAndArtifactId(ctx context.Context, groupId, artifactId string, limit int) ([]*response.Artifact, error) {
	if limit <= 0 {
		return c.IteratorByGroupAndArtifactId(ctx, groupId, artifactId).ToSlice()
	} else {
		query := request.NewQuery().SetGroupId(groupId).SetArtifactId(artifactId)
		search := request.NewSearchRequest().SetQuery(query).SetLimit(limit)
		result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, search)
		if err != nil {
			return nil, err
		}
		if result == nil || result.ResponseBody == nil {
			return nil, errors.New("empty response body")
		}
		return result.ResponseBody.Docs, nil
	}
}

// IteratorByGroupAndArtifactId 根据组ID和制品ID获取制品迭代器
func (c *Client) IteratorByGroupAndArtifactId(ctx context.Context, groupId, artifactId string) *SearchIterator[*response.Artifact] {
	query := request.NewQuery().SetGroupId(groupId).SetArtifactId(artifactId)
	search := request.NewSearchRequest().SetQuery(query)
	return NewSearchIterator[*response.Artifact](search).WithClient(c)
}

// GetArtifactDetails 获取制品的详细信息
// 如果提供了版本号，则获取特定版本的详情；否则获取最新版本的详情
// 参数:
//   - ctx: 上下文
//   - groupId: 组ID
//   - artifactId: 制品ID
//   - version: 版本号（可选，如果为空则使用最新版本）
//
// 返回:
//   - 制品详情
//   - 错误信息
func (c *Client) GetArtifactDetails(ctx context.Context, groupId, artifactId, version string) (*response.ArtifactMetadata, error) {
	// 先获取基本信息
	artifacts, err := c.SearchByGroupAndArtifactId(ctx, groupId, artifactId, 1)
	if err != nil {
		return nil, err
	}

	if len(artifacts) == 0 {
		return nil, fmt.Errorf("制品不存在: %s:%s", groupId, artifactId)
	}

	artifact := artifacts[0]

	// 如果未提供版本，使用最新版本
	if version == "" {
		version = artifact.LatestVersion
	}

	// 获取详细元数据
	return c.GetArtifactMetadata(ctx, groupId, artifactId, version)
}

// SearchPopularArtifacts 搜索热门制品
// 参数:
//   - ctx: 上下文
//   - limit: 最大返回结果数量
//
// 返回:
//   - 制品列表，按流行度排序
//   - 错误信息
func (c *Client) SearchPopularArtifacts(ctx context.Context, limit int) ([]*response.Artifact, error) {
	// 创建搜索请求，按版本数量和时间戳排序
	search := request.NewSearchRequest().
		SetQuery(request.NewQuery().SetText("*")).
		SetSort("versionCount", false). // 按版本数量降序排序
		SetLimit(limit)

	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, search)
	if err != nil {
		return nil, err
	}

	if result == nil || result.ResponseBody == nil {
		return nil, errors.New("empty response body")
	}

	return result.ResponseBody.Docs, nil
}

// SearchArtifactsByTag 根据标签搜索制品
// 参数:
//   - ctx: 上下文
//   - tag: 标签
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - 制品列表
//   - 错误信息
func (c *Client) SearchArtifactsByTag(ctx context.Context, tag string, limit int) ([]*response.Artifact, error) {
	if limit <= 0 {
		return c.IteratorByTag(ctx, tag).ToSlice()
	} else {
		query := request.NewQuery().SetTags(tag)
		search := request.NewSearchRequest().SetQuery(query).SetLimit(limit)
		result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, search)
		if err != nil {
			return nil, err
		}
		if result == nil || result.ResponseBody == nil {
			return nil, errors.New("empty response body")
		}
		return result.ResponseBody.Docs, nil
	}
}

// SearchArtifactsWithFacets 搜索制品并返回聚合结果
// 参数:
//   - ctx: 上下文
//   - searchText: 搜索文本
//   - facetFields: 要聚合的字段，如"g"表示按组ID聚合
//   - limit: 最大返回结果数量
//
// 返回:
//   - 制品列表
//   - 聚合结果
//   - 错误信息
func (c *Client) SearchArtifactsWithFacets(ctx context.Context, searchText string, facetFields []string, limit int) ([]*response.Artifact, *response.FacetResults, error) {
	// 创建搜索请求
	query := request.NewQuery().SetText(searchText)
	search := request.NewSearchRequest().
		SetQuery(query).
		SetLimit(limit).
		EnableFacet(facetFields...)

	// 执行搜索
	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, search)
	if err != nil {
		return nil, nil, err
	}

	if result == nil || result.ResponseBody == nil {
		return nil, nil, errors.New("empty response body")
	}

	// 解析聚合结果
	facetResults := &response.FacetResults{
		Counts: make(map[string][]response.FacetCount),
	}

	if result.FacetCounts != nil && result.FacetCounts.FacetFields != nil {
		for field, values := range result.FacetCounts.FacetFields {
			facetCounts := make([]response.FacetCount, 0)

			// 解析聚合值和计数
			for i := 0; i < len(values); i += 2 {
				if valueStr, ok := values[i].(string); ok {
					if countFloat, ok := values[i+1].(float64); ok {
						facetCounts = append(facetCounts, response.FacetCount{
							Value: valueStr,
							Count: int(countFloat),
						})
					}
				}
			}

			facetResults.Counts[field] = facetCounts
		}
	}

	return result.ResponseBody.Docs, facetResults, nil
}

// ArtifactDependencyInfo 制品依赖关系信息
type ArtifactDependencyInfo struct {
	// 直接依赖项
	DirectDependencies []*response.Dependency `json:"directDependencies"`

	// 传递依赖项（依赖的依赖）
	TransitiveDependencies []*response.Dependency `json:"transitiveDependencies"`

	// 可选依赖项
	OptionalDependencies []*response.Dependency `json:"optionalDependencies"`
}

// GetArtifactDependencies 获取制品的依赖关系
// 参数:
//   - ctx: 上下文
//   - groupId: 组ID
//   - artifactId: 制品ID
//   - version: 版本号
//
// 返回:
//   - 制品依赖关系信息
//   - 错误信息
func (c *Client) GetArtifactDependencies(ctx context.Context, groupId, artifactId, version string) (*ArtifactDependencyInfo, error) {
	// 获取制品元数据
	metadata, err := c.GetArtifactMetadata(ctx, groupId, artifactId, version)
	if err != nil {
		return nil, err
	}

	// 分类依赖项
	depInfo := &ArtifactDependencyInfo{
		DirectDependencies:     make([]*response.Dependency, 0),
		OptionalDependencies:   make([]*response.Dependency, 0),
		TransitiveDependencies: make([]*response.Dependency, 0),
	}

	// 处理依赖项
	for _, dep := range metadata.Dependencies {
		if dep.Optional {
			depInfo.OptionalDependencies = append(depInfo.OptionalDependencies, dep)
		} else if dep.Scope == "compile" || dep.Scope == "runtime" {
			depInfo.DirectDependencies = append(depInfo.DirectDependencies, dep)
		} else {
			depInfo.TransitiveDependencies = append(depInfo.TransitiveDependencies, dep)
		}
	}

	return depInfo, nil
}

// ArtifactUsage 制品使用情况
type ArtifactUsage struct {
	// 总使用者数量
	TotalUsageCount int `json:"totalUsageCount"`

	// 使用此制品的前N个项目
	TopUsers []*response.Artifact `json:"topUsers"`

	// 按组ID分组的使用者数量
	UsageByGroup map[string]int `json:"usageByGroup"`
}

// GetArtifactUsage 获取制品的使用情况
// 参数:
//   - ctx: 上下文
//   - groupId: 组ID
//   - artifactId: 制品ID
//   - limit: 返回的顶级使用者数量
//
// 返回:
//   - 制品使用情况
//   - 错误信息
func (c *Client) GetArtifactUsage(ctx context.Context, groupId, artifactId, version string, limit int) (*ArtifactUsage, error) {
	// 构建搜索查询
	dependencyQuery := fmt.Sprintf("d:%s:%s", groupId, artifactId)
	if version != "" {
		dependencyQuery += ":" + version
	}

	// 创建搜索请求
	query := request.NewQuery().SetCustomQuery(dependencyQuery)
	search := request.NewSearchRequest().SetQuery(query).SetLimit(limit)

	// 执行搜索
	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, search)
	if err != nil {
		return nil, err
	}

	if result == nil || result.ResponseBody == nil {
		return nil, errors.New("empty response body")
	}

	// 初始化使用情况对象
	usage := &ArtifactUsage{
		TotalUsageCount: result.ResponseBody.NumFound,
		TopUsers:        result.ResponseBody.Docs,
		UsageByGroup:    make(map[string]int),
	}

	// 按组统计使用情况
	for _, artifact := range usage.TopUsers {
		usage.UsageByGroup[artifact.GroupId] = usage.UsageByGroup[artifact.GroupId] + 1
	}

	return usage, nil
}

// ArtifactComparisonResult 制品比较结果
type ArtifactComparisonResult struct {
	// 基本信息
	Artifact1 *response.Artifact `json:"artifact1"`
	Artifact2 *response.Artifact `json:"artifact2"`

	// 版本数量差异
	VersionCountDiff int `json:"versionCountDiff"`

	// 活跃度比较（基于最新更新时间和版本数量）
	MostActive string `json:"mostActive"`

	// 流行度比较
	MostPopular string `json:"mostPopular"`

	// 更新时间差异（天）
	UpdateTimeDiffDays int `json:"updateTimeDiffDays"`
}

// CompareArtifacts 比较两个制品
// 参数:
//   - ctx: 上下文
//   - groupId1: 第一个制品的组ID
//   - artifactId1: 第一个制品的制品ID
//   - groupId2: 第二个制品的组ID
//   - artifactId2: 第二个制品的制品ID
//
// 返回:
//   - 比较结果
//   - 错误信息
func (c *Client) CompareArtifacts(ctx context.Context, groupId1, artifactId1, groupId2, artifactId2 string) (*ArtifactComparisonResult, error) {
	// 获取第一个制品
	artifacts1, err := c.SearchByGroupAndArtifactId(ctx, groupId1, artifactId1, 1)
	if err != nil {
		return nil, err
	}
	if len(artifacts1) == 0 {
		return nil, fmt.Errorf("制品不存在: %s:%s", groupId1, artifactId1)
	}

	// 获取第二个制品
	artifacts2, err := c.SearchByGroupAndArtifactId(ctx, groupId2, artifactId2, 1)
	if err != nil {
		return nil, err
	}
	if len(artifacts2) == 0 {
		return nil, fmt.Errorf("制品不存在: %s:%s", groupId2, artifactId2)
	}

	artifact1 := artifacts1[0]
	artifact2 := artifacts2[0]

	// 计算版本数量差异
	versionCountDiff := artifact1.VersionCount - artifact2.VersionCount

	// 计算更新时间差异
	time1 := time.Unix(artifact1.Timestamp/1000, 0)
	time2 := time.Unix(artifact2.Timestamp/1000, 0)
	timeDiff := time1.Sub(time2)
	updateTimeDiffDays := int(timeDiff.Hours() / 24)

	// 确定最活跃的制品
	var mostActive string
	if artifact1.VersionCount > artifact2.VersionCount && time1.After(time2) {
		mostActive = fmt.Sprintf("%s:%s", groupId1, artifactId1)
	} else if artifact2.VersionCount > artifact1.VersionCount && time2.After(time1) {
		mostActive = fmt.Sprintf("%s:%s", groupId2, artifactId2)
	} else {
		// 如果一个指标更高，另一个更低，根据综合计算
		score1 := float64(artifact1.VersionCount) * (float64(artifact1.Timestamp) / 1000000)
		score2 := float64(artifact2.VersionCount) * (float64(artifact2.Timestamp) / 1000000)

		if score1 > score2 {
			mostActive = fmt.Sprintf("%s:%s", groupId1, artifactId1)
		} else {
			mostActive = fmt.Sprintf("%s:%s", groupId2, artifactId2)
		}
	}

	// 确定最流行的制品
	var mostPopular string
	if artifact1.VersionCount > artifact2.VersionCount {
		mostPopular = fmt.Sprintf("%s:%s", groupId1, artifactId1)
	} else if artifact2.VersionCount > artifact1.VersionCount {
		mostPopular = fmt.Sprintf("%s:%s", groupId2, artifactId2)
	} else {
		// 版本数相同，根据更新时间判断
		if time1.After(time2) {
			mostPopular = fmt.Sprintf("%s:%s", groupId1, artifactId1)
		} else {
			mostPopular = fmt.Sprintf("%s:%s", groupId2, artifactId2)
		}
	}

	return &ArtifactComparisonResult{
		Artifact1:          artifact1,
		Artifact2:          artifact2,
		VersionCountDiff:   versionCountDiff,
		MostActive:         mostActive,
		MostPopular:        mostPopular,
		UpdateTimeDiffDays: updateTimeDiffDays,
	}, nil
}

// SearchArtifactsByDateRange 根据日期范围搜索制品
// 参数:
//   - ctx: 上下文
//   - startDate: 开始日期，格式为YYYY-MM-DD
//   - endDate: 结束日期，格式为YYYY-MM-DD
//   - limit: 最大返回结果数量
//
// 返回:
//   - 制品列表
//   - 错误信息
func (c *Client) SearchArtifactsByDateRange(ctx context.Context, startDate, endDate string, limit int) ([]*response.Artifact, error) {
	// 构建日期范围查询
	dateQuery := fmt.Sprintf("timestamp:[%s TO %s]", startDate, endDate)

	// 创建搜索请求
	query := request.NewQuery().SetCustomQuery(dateQuery)
	search := request.NewSearchRequest().
		SetQuery(query).
		SetSort("timestamp", false). // 按时间戳降序排序
		SetLimit(limit)

	// 执行搜索
	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, search)
	if err != nil {
		return nil, err
	}

	if result == nil || result.ResponseBody == nil {
		return nil, errors.New("empty response body")
	}

	return result.ResponseBody.Docs, nil
}

// SuggestSimilarArtifacts 根据指定制品推荐相似制品
// 参数:
//   - ctx: 上下文
//   - groupId: 组ID
//   - artifactId: 制品ID
//   - limit: 最大返回结果数量
//
// 返回:
//   - 推荐的相似制品列表
//   - 错误信息
func (c *Client) SuggestSimilarArtifacts(ctx context.Context, groupId, artifactId string, limit int) ([]*response.Artifact, error) {
	// 步骤1: 获取目标制品的详情
	artifacts, err := c.SearchByGroupAndArtifactId(ctx, groupId, artifactId, 1)
	if err != nil {
		return nil, err
	}

	if len(artifacts) == 0 {
		return nil, fmt.Errorf("制品不存在: %s:%s", groupId, artifactId)
	}

	baseArtifact := artifacts[0]

	// 步骤2: 从目标制品的标签和文本分析关键词
	keywords := make([]string, 0)

	// 从标签中提取关键词
	keywords = append(keywords, baseArtifact.Tags...)

	// 从文本中提取可能的关键词
	for _, text := range baseArtifact.Text {
		// 简单处理，按空格分割
		parts := strings.Split(text, " ")
		for _, part := range parts {
			if len(part) > 3 && !contains(keywords, part) {
				keywords = append(keywords, part)
			}
		}
	}

	// 如果关键词太少，添加artifactId作为关键词
	if len(keywords) < 2 {
		keywords = append(keywords, artifactId)
	}

	// 步骤3: 构建搜索查询，排除自身
	query := fmt.Sprintf("NOT (g:%s AND a:%s)", groupId, artifactId)

	// 添加关键词，限制最多使用5个关键词
	keywordLimit := 5
	if len(keywords) > keywordLimit {
		keywords = keywords[:keywordLimit]
	}

	if len(keywords) > 0 {
		keywordQuery := strings.Join(keywords, " OR ")
		query = fmt.Sprintf("(%s) AND (%s)", query, keywordQuery)
	}

	// 步骤4: 执行搜索
	search := request.NewSearchRequest().
		SetQuery(request.NewQuery().SetCustomQuery(query)).
		SetLimit(limit)

	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, search)
	if err != nil {
		return nil, err
	}

	if result == nil || result.ResponseBody == nil {
		return nil, errors.New("empty response body")
	}

	return result.ResponseBody.Docs, nil
}

// 辅助函数：检查字符串切片是否包含指定字符串
func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// ArtifactStats 制品统计信息
type ArtifactStats struct {
	// 基本信息
	GroupId    string `json:"groupId"`
	ArtifactId string `json:"artifactId"`

	// 版本统计
	TotalVersions     int   `json:"totalVersions"`
	LatestVersionDate int64 `json:"latestVersionDate"`
	FirstVersionDate  int64 `json:"firstVersionDate"`

	// 活跃度指标
	DaysSinceLastUpdate int     `json:"daysSinceLastUpdate"`
	UpdateFrequency     float64 `json:"updateFrequency"` // 平均每月发布版本数

	// 流行度指标
	UsageCount int `json:"usageCount"` // 被其他制品依赖的次数
}

// GetArtifactStats 获取制品的统计信息
// 参数:
//   - ctx: 上下文
//   - groupId: 组ID
//   - artifactId: 制品ID
//
// 返回:
//   - 制品统计信息
//   - 错误信息
func (c *Client) GetArtifactStats(ctx context.Context, groupId, artifactId string) (*ArtifactStats, error) {
	// 获取制品基本信息
	artifacts, err := c.SearchByGroupAndArtifactId(ctx, groupId, artifactId, 1)
	if err != nil {
		return nil, err
	}

	if len(artifacts) == 0 {
		return nil, fmt.Errorf("制品不存在: %s:%s", groupId, artifactId)
	}

	// 注意：此变量在当前实现中未使用
	// 但当需要使用此变量获取更多信息时可以取消注释
	// artifact := artifacts[0]

	// 获取所有版本信息
	versions, err := c.ListVersions(ctx, groupId, artifactId, 0)
	if err != nil {
		return nil, err
	}

	// 初始化统计信息
	stats := &ArtifactStats{
		GroupId:       groupId,
		ArtifactId:    artifactId,
		TotalVersions: len(versions),
	}

	// 如果有版本信息，计算时间相关指标
	if len(versions) > 0 {
		// 按时间戳排序
		sort.Slice(versions, func(i, j int) bool {
			return versions[i].Timestamp > versions[j].Timestamp
		})

		// 最新版本和首个版本的日期
		stats.LatestVersionDate = versions[0].Timestamp
		stats.FirstVersionDate = versions[len(versions)-1].Timestamp

		// 计算距最后更新的天数
		lastUpdateTime := time.Unix(stats.LatestVersionDate/1000, 0)
		stats.DaysSinceLastUpdate = int(time.Since(lastUpdateTime).Hours() / 24)

		// 计算更新频率（平均每月发布版本数）
		if stats.TotalVersions > 1 {
			totalMonths := float64(stats.LatestVersionDate-stats.FirstVersionDate) / 1000 / 60 / 60 / 24 / 30
			if totalMonths > 0 {
				stats.UpdateFrequency = float64(stats.TotalVersions) / totalMonths
			}
		}
	}

	// 获取使用情况
	usage, err := c.GetArtifactUsage(ctx, groupId, artifactId, "", 0)
	if err == nil {
		stats.UsageCount = usage.TotalUsageCount
	}

	return stats, nil
}
