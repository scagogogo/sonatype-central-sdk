package api

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchByTag 根据标签搜索项目
func (c *Client) SearchByTag(ctx context.Context, tag string, limit int) ([]*response.Artifact, error) {
	if limit <= 0 {
		return c.IteratorByTag(ctx, tag).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetTags(tag)).SetLimit(limit)
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

// IteratorByTag 返回根据标签搜索的迭代器
func (c *Client) IteratorByTag(ctx context.Context, tag string) *SearchIterator[*response.Artifact] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetTags(tag))
	return NewSearchIterator[*response.Artifact](search).WithClient(c)
}

// SearchByMultipleTags 搜索同时具有多个标签的项目
func (c *Client) SearchByMultipleTags(ctx context.Context, tags []string, limit int) ([]*response.Artifact, error) {
	if len(tags) == 0 {
		return nil, errors.New("at least one tag must be provided")
	}

	// 创建自定义查询字符串来包含所有标签
	var queryParts []string
	for _, tag := range tags {
		queryParts = append(queryParts, "tags:"+tag)
	}

	// 使用AND连接所有标签查询条件
	customQuery := strings.Join(queryParts, " AND ")

	// 使用自定义查询
	query := request.NewQuery().SetCustomQuery(customQuery)

	if limit <= 0 {
		searchRequest := request.NewSearchRequest().SetQuery(query)
		iterator := NewSearchIterator[*response.Artifact](searchRequest).WithClient(c)
		return iterator.ToSlice()
	} else {
		searchRequest := request.NewSearchRequest().SetQuery(query).SetLimit(limit)
		result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, searchRequest)
		if err != nil {
			return nil, err
		}
		if result == nil || result.ResponseBody == nil {
			return nil, errors.New("empty response body")
		}
		return result.ResponseBody.Docs, nil
	}
}

// GetMostUsedTags 获取最常用的标签
func (c *Client) GetMostUsedTags(ctx context.Context, baseTag string, limit int) ([]response.TagCount, error) {
	// 通过基础标签查询常见项目，如查询"java"项目
	var artifacts []*response.Artifact
	var err error

	if baseTag != "" {
		artifacts, err = c.SearchByTag(ctx, baseTag, 200) // 获取足够多的样本
	} else {
		// 如果没有指定基础标签，尝试获取一些热门的依赖项
		popularGroupIds := []string{"org.springframework", "com.google.guava", "org.apache.commons"}
		for _, groupId := range popularGroupIds {
			someArtifacts, err := c.SearchByGroupId(ctx, groupId, 50)
			if err == nil && len(someArtifacts) > 0 {
				artifacts = append(artifacts, someArtifacts...)
			}
			if len(artifacts) >= 200 {
				break
			}
		}
	}

	if err != nil {
		return nil, err
	}

	if len(artifacts) == 0 {
		return nil, fmt.Errorf("no artifacts found to analyze for tags")
	}

	// 统计标签
	tagCounts := make(map[string]int)
	for _, artifact := range artifacts {
		for _, tag := range artifact.Tags {
			tagCounts[tag]++
		}
	}

	// 转换为结果集并排序
	var results []response.TagCount
	for tag, count := range tagCounts {
		results = append(results, response.TagCount{
			Tag:   tag,
			Count: count,
		})
	}

	// 排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Count > results[j].Count
	})

	// 限制结果数量
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// SearchArtifactsWithAllTags 搜索同时拥有所有指定标签的项目
func (c *Client) SearchArtifactsWithAllTags(ctx context.Context, tags []string, limit int) ([]*response.Artifact, error) {
	if len(tags) == 0 {
		return nil, errors.New("at least one tag must be provided")
	}

	// 首先获取第一个标签的结果
	artifacts, err := c.SearchByTag(ctx, tags[0], 0) // 获取所有结果
	if err != nil {
		return nil, err
	}

	// 如果只有一个标签，直接返回
	if len(tags) == 1 {
		if limit > 0 && len(artifacts) > limit {
			return artifacts[:limit], nil
		}
		return artifacts, nil
	}

	// 过滤那些包含所有标签的项目
	var filteredArtifacts []*response.Artifact
	for _, artifact := range artifacts {
		hasAllTags := true
		for _, requiredTag := range tags[1:] { // 跳过第一个标签，因为已经用它进行了初始搜索
			if !containsTag(artifact.Tags, requiredTag) {
				hasAllTags = false
				break
			}
		}
		if hasAllTags {
			filteredArtifacts = append(filteredArtifacts, artifact)
			if limit > 0 && len(filteredArtifacts) >= limit {
				break
			}
		}
	}

	return filteredArtifacts, nil
}

// SearchByTagWithGroupFilter 根据标签搜索并按GroupId过滤
func (c *Client) SearchByTagWithGroupFilter(ctx context.Context, tag string, groupIdPrefix string, limit int) ([]*response.Artifact, error) {
	// 获取标签的所有结果
	artifacts, err := c.SearchByTag(ctx, tag, 0)
	if err != nil {
		return nil, err
	}

	// 过滤GroupId
	var filteredArtifacts []*response.Artifact
	for _, artifact := range artifacts {
		if strings.HasPrefix(artifact.GroupId, groupIdPrefix) {
			filteredArtifacts = append(filteredArtifacts, artifact)
			if limit > 0 && len(filteredArtifacts) >= limit {
				break
			}
		}
	}

	return filteredArtifacts, nil
}

// 辅助函数：检查标签数组中是否包含指定标签
func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// CountArtifactsByTag 计算具有特定标签的构件数量
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - tag: 要计数的标签
//
// 返回:
//   - int: 使用此标签的构件数量
//   - error: 如果计数过程中发生错误
func (c *Client) CountArtifactsByTag(ctx context.Context, tag string) (int, error) {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetTags(tag)).SetLimit(0)
	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, search)
	if err != nil {
		return 0, err
	}
	if result == nil || result.ResponseBody == nil {
		return 0, errors.New("empty response body")
	}
	return result.ResponseBody.NumFound, nil
}

// CountVersions 计算指定构件的版本数量
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - groupId: 构件的groupId
//   - artifactId: 构件的artifactId
//
// 返回:
//   - int: 版本数量
//   - error: 如果计数过程中发生错误
func (c *Client) CountVersions(ctx context.Context, groupId, artifactId string) (int, error) {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId(groupId).SetArtifactId(artifactId)).SetCore("gav").SetLimit(0)
	result, err := SearchRequestJsonDoc[*response.Version](c, ctx, search)
	if err != nil {
		return 0, err
	}
	if result == nil || result.ResponseBody == nil {
		return 0, errors.New("empty response body")
	}
	return result.ResponseBody.NumFound, nil
}

// SearchByTagAndSortByPopularity 根据标签搜索并按流行度排序
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - tag: 要搜索的标签
//   - limit: 最大返回结果数，如果小于等于0则返回所有结果
//
// 返回:
//   - []*response.Artifact: 按流行度排序的构件列表
//   - error: 如果搜索过程中发生错误
func (c *Client) SearchByTagAndSortByPopularity(ctx context.Context, tag string, limit int) ([]*response.Artifact, error) {
	artifacts, err := c.SearchByTag(ctx, tag, 0)
	if err != nil {
		return nil, err
	}

	// 计算每个构件的"受欢迎度"（根据下载量或使用次数，这里简化为版本数量）
	type ArtifactWithPopularity struct {
		Artifact *response.Artifact
		Score    int
	}

	var scoredArtifacts []ArtifactWithPopularity

	// 获取每个构件的版本数量作为流行度指标
	for _, artifact := range artifacts {
		// 此处可以根据需要替换为其他流行度指标
		versionCount, err := c.CountVersions(ctx, artifact.GroupId, artifact.ArtifactId)
		if err == nil && versionCount > 0 {
			scoredArtifacts = append(scoredArtifacts, ArtifactWithPopularity{
				Artifact: artifact,
				Score:    versionCount,
			})
		}
	}

	// 按流行度分数排序（降序）
	sort.Slice(scoredArtifacts, func(i, j int) bool {
		return scoredArtifacts[i].Score > scoredArtifacts[j].Score
	})

	// 提取排序后的构件
	var result []*response.Artifact
	for _, scoredArtifact := range scoredArtifacts {
		result = append(result, scoredArtifact.Artifact)
	}

	// 限制结果数量
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}

// SearchByTagPrefix 使用标签前缀进行模糊搜索
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - prefix: 标签前缀
//   - limit: 最大返回结果数，如果小于等于0则返回所有结果
//
// 返回:
//   - []*response.Artifact: 匹配标签前缀的构件列表
//   - error: 如果搜索过程中发生错误
func (c *Client) SearchByTagPrefix(ctx context.Context, prefix string, limit int) ([]*response.Artifact, error) {
	if prefix == "" {
		return nil, errors.New("tag prefix cannot be empty")
	}

	// 使用自定义查询构建标签前缀搜索
	customQuery := "tags:" + prefix + "*"
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetCustomQuery(customQuery))

	if limit > 0 {
		search.SetLimit(limit)
	}

	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, search)
	if err != nil {
		return nil, err
	}
	if result == nil || result.ResponseBody == nil {
		return nil, errors.New("empty response body")
	}
	return result.ResponseBody.Docs, nil
}
