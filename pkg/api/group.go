package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchByGroupId 根据GroupID列出这个组下面的artifact
func (c *Client) SearchByGroupId(ctx context.Context, groupId string, limit int) ([]*response.Artifact, error) {
	if limit <= 0 {
		return c.IteratorByGroupId(ctx, groupId).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId(groupId)).SetLimit(limit)
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

func (c *Client) IteratorByGroupId(ctx context.Context, groupId string) *SearchIterator[*response.Artifact] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId(groupId))
	return NewSearchIterator[*response.Artifact](search).WithClient(c)
}

// SearchByGroupPattern 根据模式（如前缀、关键词等）搜索组ID
func (c *Client) SearchByGroupPattern(ctx context.Context, pattern string, limit int) ([]*response.GroupSearchResult, error) {
	// 构建查询 - 注意这里使用了g开头的模糊匹配搜索
	q := fmt.Sprintf("g:%s*", pattern)
	query := request.NewQuery().SetCustomQuery(q)
	searchReq := request.NewSearchRequest().
		SetQuery(query).
		SetRows(limit)

	// 获取结果
	var result response.Response[map[string]interface{}]
	err := c.SearchRequest(ctx, searchReq, &result)
	if err != nil {
		return nil, fmt.Errorf("搜索组ID模式失败: %w", err)
	}

	if result.ResponseBody.NumFound == 0 {
		return []*response.GroupSearchResult{}, nil
	}

	// 处理结果 - 提取唯一的groupId
	groupMap := make(map[string]*response.GroupSearchResult)
	for _, doc := range result.ResponseBody.Docs {
		groupId, ok := doc["g"].(string)
		if !ok || groupId == "" {
			continue
		}

		artifactId, _ := doc["a"].(string)
		version, _ := doc["v"].(string)
		timestamp, _ := doc["timestamp"].(float64)

		// 如果这个组ID已经存在，就更新信息
		if group, exists := groupMap[groupId]; exists {
			group.ArtifactCount++
			if timestamp > group.LastUpdated {
				group.LastUpdated = timestamp
				group.LastUpdatedDate = time.UnixMilli(int64(timestamp)).Format(time.RFC3339)
			}
			group.Artifacts = append(group.Artifacts, &response.GroupArtifact{
				ArtifactId: artifactId,
				Version:    version,
			})
		} else {
			// 否则创建新的组记录
			groupMap[groupId] = &response.GroupSearchResult{
				GroupId:         groupId,
				ArtifactCount:   1,
				LastUpdated:     timestamp,
				LastUpdatedDate: time.UnixMilli(int64(timestamp)).Format(time.RFC3339),
				Artifacts: []*response.GroupArtifact{
					{
						ArtifactId: artifactId,
						Version:    version,
					},
				},
			}
		}
	}

	// 将map转换为slice
	groups := make([]*response.GroupSearchResult, 0, len(groupMap))
	for _, group := range groupMap {
		groups = append(groups, group)
	}

	return groups, nil
}

// GetGroupStatistics 获取组的统计信息（如组内artifact数量、版本数量等）
func (c *Client) GetGroupStatistics(ctx context.Context, groupId string) (*response.GroupStatistics, error) {
	// 首先获取该组下的所有artifact
	artifacts, err := c.SearchByGroupId(ctx, groupId, 0) // 0表示获取所有
	if err != nil {
		return nil, fmt.Errorf("获取组统计信息失败: %w", err)
	}

	if len(artifacts) == 0 {
		return nil, fmt.Errorf("组 %s 不存在或没有artifacts", groupId)
	}

	// 准备统计信息
	stats := &response.GroupStatistics{
		GroupId:       groupId,
		ArtifactCount: len(artifacts),
		Artifacts:     make([]*response.ArtifactStatistics, 0, len(artifacts)),
	}

	var totalVersions int
	var latestUpdate int64

	// 遍历每个artifact，获取它的版本信息
	for _, artifact := range artifacts {
		// 获取这个artifact的所有版本
		versions, err := c.ListVersions(ctx, groupId, artifact.ArtifactId, 0)
		if err != nil {
			// 如果获取某个artifact的版本失败，我们继续处理其他的
			continue
		}

		// 更新统计信息
		totalVersions += len(versions)

		// 找出最新的更新时间
		for _, version := range versions {
			if version.Timestamp > latestUpdate {
				latestUpdate = version.Timestamp
				stats.LastUpdatedDate = time.UnixMilli(version.Timestamp).Format(time.RFC3339)
			}
		}

		// 添加这个artifact的统计信息
		artifactStats := &response.ArtifactStatistics{
			ArtifactId:    artifact.ArtifactId,
			VersionCount:  len(versions),
			LatestVersion: "", // 暂时设为空字符串
		}
		stats.Artifacts = append(stats.Artifacts, artifactStats)
	}

	stats.TotalVersions = totalVersions
	stats.LatestUpdate = latestUpdate

	return stats, nil
}

// GetPopularGroups 获取流行的组（按使用频率排序）
func (c *Client) GetPopularGroups(ctx context.Context, limit int) ([]*response.GroupPopularity, error) {
	// 使用facet查询获取组分布
	query := request.NewQuery().SetCustomQuery("*:*")
	searchReq := request.NewSearchRequest().
		SetQuery(query).
		AddCustomParam("facet", "true").
		AddCustomParam("facet.field", "g").
		AddCustomParam("facet.limit", fmt.Sprintf("%d", limit)).
		SetRows(0) // 只需要聚合结果，不需要文档

	// 执行查询
	var result response.Response[map[string]interface{}]
	err := c.SearchRequest(ctx, searchReq, &result)
	if err != nil {
		return nil, fmt.Errorf("获取流行组失败: %w", err)
	}

	// 处理facet结果
	groups := make([]*response.GroupPopularity, 0)

	if result.FacetCounts != nil && result.FacetCounts.FacetFields != nil {
		if groupField, ok := result.FacetCounts.FacetFields["g"]; ok {
			// facet结果格式为[group1, count1, group2, count2, ...]
			for i := 0; i < len(groupField); i += 2 {
				if groupName, ok := groupField[i].(string); ok {
					if count, ok := groupField[i+1].(float64); ok {
						groups = append(groups, &response.GroupPopularity{
							GroupId:        groupName,
							ArtifactCount:  int(count),
							PopularityRank: i/2 + 1,
						})
					}
				}
			}
		}
	}

	return groups, nil
}

// CompareTwoGroups 比较两个组的基本信息和统计数据
func (c *Client) CompareTwoGroups(ctx context.Context, groupId1, groupId2 string) (*response.GroupComparison, error) {
	// 获取两个组的统计信息
	stats1, err1 := c.GetGroupStatistics(ctx, groupId1)
	stats2, err2 := c.GetGroupStatistics(ctx, groupId2)

	// 创建比较结果
	comparison := &response.GroupComparison{
		Group1:      groupId1,
		Group2:      groupId2,
		Group1Error: "",
		Group2Error: "",
	}

	// 处理错误情况
	if err1 != nil {
		comparison.Group1Error = err1.Error()
	} else {
		comparison.Group1Stats = stats1
	}

	if err2 != nil {
		comparison.Group2Error = err2.Error()
	} else {
		comparison.Group2Stats = stats2
	}

	// 如果两个组都获取失败，返回错误
	if err1 != nil && err2 != nil {
		return comparison, fmt.Errorf("两个组都获取失败: %s, %s", err1.Error(), err2.Error())
	}

	// 计算共同的artifacts（如果两个组都获取成功）
	if err1 == nil && err2 == nil {
		// 创建GroupId1的artifact映射
		artifactMap1 := make(map[string]bool)
		for _, artifact := range stats1.Artifacts {
			artifactMap1[artifact.ArtifactId] = true
		}

		// 查找共同的artifacts
		var commonArtifacts []string
		for _, artifact := range stats2.Artifacts {
			if artifactMap1[artifact.ArtifactId] {
				commonArtifacts = append(commonArtifacts, artifact.ArtifactId)
			}
		}

		comparison.CommonArtifacts = commonArtifacts
		comparison.CommonArtifactCount = len(commonArtifacts)
	}

	return comparison, nil
}

// SearchSubgroups 搜索一个组的所有子组
func (c *Client) SearchSubgroups(ctx context.Context, parentGroupId string, limit int) ([]*response.GroupSearchResult, error) {
	// 确保parentGroupId以点号结尾，用于搜索子组
	if !strings.HasSuffix(parentGroupId, ".") {
		parentGroupId = parentGroupId + "."
	}

	// 构建查询 - 搜索以parentGroupId开头的所有groupId
	q := fmt.Sprintf("g:%s*", parentGroupId)
	query := request.NewQuery().SetCustomQuery(q)
	searchReq := request.NewSearchRequest().
		SetQuery(query).
		SetRows(limit)

	// 获取结果
	var result response.Response[map[string]interface{}]
	err := c.SearchRequest(ctx, searchReq, &result)
	if err != nil {
		return nil, fmt.Errorf("搜索子组失败: %w", err)
	}

	if result.ResponseBody.NumFound == 0 {
		return []*response.GroupSearchResult{}, nil
	}

	// 处理结果 - 提取唯一的groupId，且必须是parentGroupId的直接子组
	groupMap := make(map[string]*response.GroupSearchResult)
	for _, doc := range result.ResponseBody.Docs {
		groupId, ok := doc["g"].(string)
		if !ok || groupId == "" {
			continue
		}

		// 跳过父组本身
		if groupId == strings.TrimSuffix(parentGroupId, ".") {
			continue
		}

		// 确保是直接子组，而不是孙子组或更深层次
		subGroup := strings.TrimPrefix(groupId, strings.TrimSuffix(parentGroupId, "."))
		if strings.Contains(subGroup, ".") {
			// 提取第一级子组
			parts := strings.SplitN(subGroup, ".", 2)
			subGroup = strings.TrimSuffix(parentGroupId, ".") + "." + parts[0]
		} else {
			subGroup = groupId
		}

		artifactId, _ := doc["a"].(string)
		version, _ := doc["v"].(string)
		timestamp, _ := doc["timestamp"].(float64)

		// 如果这个子组ID已经存在，就更新信息
		if group, exists := groupMap[subGroup]; exists {
			group.ArtifactCount++
			if timestamp > group.LastUpdated {
				group.LastUpdated = timestamp
				group.LastUpdatedDate = time.UnixMilli(int64(timestamp)).Format(time.RFC3339)
			}
			// 只添加不同的artifact
			artifactExists := false
			for _, a := range group.Artifacts {
				if a.ArtifactId == artifactId {
					artifactExists = true
					break
				}
			}
			if !artifactExists {
				group.Artifacts = append(group.Artifacts, &response.GroupArtifact{
					ArtifactId: artifactId,
					Version:    version,
				})
			}
		} else {
			// 否则创建新的组记录
			groupMap[subGroup] = &response.GroupSearchResult{
				GroupId:         subGroup,
				ArtifactCount:   1,
				LastUpdated:     timestamp,
				LastUpdatedDate: time.UnixMilli(int64(timestamp)).Format(time.RFC3339),
				Artifacts: []*response.GroupArtifact{
					{
						ArtifactId: artifactId,
						Version:    version,
					},
				},
			}
		}
	}

	// 将map转换为slice
	groups := make([]*response.GroupSearchResult, 0, len(groupMap))
	for _, group := range groupMap {
		groups = append(groups, group)
	}

	return groups, nil
}

// GetGroupInfo 获取关于特定groupId的基本信息
func (c *Client) GetGroupInfo(ctx context.Context, groupId string) (*response.GroupInfo, error) {
	// 首先获取该组下的所有artifact
	artifacts, err := c.SearchByGroupId(ctx, groupId, 0) // 0表示获取所有
	if err != nil {
		return nil, fmt.Errorf("获取组信息失败: %w", err)
	}

	if len(artifacts) == 0 {
		// 如果没有找到artifacts，返回一个基本信息
		return &response.GroupInfo{
			GroupId:       groupId,
			ArtifactCount: 0,
		}, nil
	}

	// 准备基本信息
	info := &response.GroupInfo{
		GroupId:       groupId,
		ArtifactCount: len(artifacts),
	}

	var latestUpdate int64

	// 获取最新更新时间
	for _, artifact := range artifacts {
		// 尝试从artifacts中获取最新的时间戳
		if artifact.Timestamp > latestUpdate {
			latestUpdate = artifact.Timestamp
		}
	}

	// 如果没有从artifacts中获取到足够的时间信息，获取版本信息
	if latestUpdate == 0 && len(artifacts) > 0 {
		versions, err := c.ListVersions(ctx, groupId, artifacts[0].ArtifactId, 0)
		if err == nil && len(versions) > 0 {
			for _, version := range versions {
				if version.Timestamp > latestUpdate {
					latestUpdate = version.Timestamp
				}
			}
		}
	}

	// 设置最后更新时间
	info.LastUpdated = latestUpdate
	if latestUpdate > 0 {
		info.LastUpdatedDate = time.UnixMilli(latestUpdate).Format(time.RFC3339)
	}

	return info, nil
}
