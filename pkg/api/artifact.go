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

// SearchByArtifactId 根据ArtifactId搜索制品
//
// 该方法提供按照ArtifactId搜索Maven制品的功能，可以获取所有匹配指定ArtifactId的制品列表。
// 当需要查找特定名称的制品，但不确定其所属的组(GroupId)时，此方法特别有用。
// 搜索结果会返回所有包含此ArtifactId的制品，可能来自不同的组织或项目。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - artifactId: 要搜索的制品ID，如"guava"、"junit"等
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - []*response.Artifact: 匹配制品的列表
//   - error: 搜索过程中发生的错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 搜索所有与"log4j"相关的制品，限制返回前10个结果
//	artifacts, err := client.SearchByArtifactId(ctx, "log4j", 10)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 处理搜索结果
//	fmt.Printf("找到 %d 个结果\n", len(artifacts))
//	for _, artifact := range artifacts {
//	    fmt.Printf("%s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
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

// IteratorByArtifactId 根据制品ID获取制品迭代器
//
// 该方法提供了一个迭代器，用于高效地分页获取并处理所有与指定ArtifactId匹配的制品。
// 当可能的搜索结果较大或需要批量处理时，迭代器模式比一次性获取所有结果更高效，
// 可以减少内存占用并提高应用响应性。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - artifactId: 要搜索的制品ID，如"spring-core"
//
// 返回:
//   - *SearchIterator[*response.Artifact]: 搜索结果的迭代器，可用于逐页获取结果
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建迭代器
//	iterator := client.IteratorByArtifactId(ctx, "guava")
//
//	// 设置每页大小
//	iterator.SetPageSize(20)
//
//	// 使用迭代器处理所有结果
//	for iterator.HasNext() {
//	    artifacts, err := iterator.Next()
//	    if err != nil {
//	        log.Fatalf("获取下一批结果失败: %v", err)
//	    }
//
//	    for _, artifact := range artifacts {
//	        fmt.Printf("%s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	    }
//	}
func (c *Client) IteratorByArtifactId(ctx context.Context, artifactId string) *SearchIterator[*response.Artifact] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetArtifactId(artifactId))
	return NewSearchIterator[*response.Artifact](search).WithClient(c)
}

// SearchByGroupAndArtifactId 根据GroupId和ArtifactId搜索制品
//
// 该方法提供更精确的搜索功能，通过同时指定GroupId和ArtifactId来查找特定的Maven制品。
// 与仅使用ArtifactId搜索相比，此方法可以更准确地定位到特定组织或项目的制品。
// 搜索结果将只包含同时匹配指定GroupId和ArtifactId的制品。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 要搜索的组ID，如"org.apache.logging.log4j"
//   - artifactId: 要搜索的制品ID，如"log4j-core"
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - []*response.Artifact: 匹配制品的列表
//   - error: 搜索过程中发生的错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 搜索Apache Commons Lang库
//	artifacts, err := client.SearchByGroupAndArtifactId(ctx, "org.apache.commons", "commons-lang3", 5)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 输出搜索结果
//	fmt.Printf("找到 %d 个制品:\n", len(artifacts))
//	for _, artifact := range artifacts {
//	    fmt.Printf("%s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
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
//
// 该方法提供了一个强大的迭代器，用于高效分页处理根据GroupId和ArtifactId查询的大量搜索结果。
// 当可能返回的结果集较大或需要批量处理时，迭代器模式比一次性加载所有结果更高效，
// 可以有效控制内存使用并提高应用程序响应性。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 要搜索的组ID，如"org.apache.commons"
//   - artifactId: 要搜索的制品ID，如"commons-lang3"
//
// 返回:
//   - *SearchIterator[*response.Artifact]: 搜索结果的迭代器，可用于逐页获取搜索结果
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建迭代器
//	iterator := client.IteratorByGroupAndArtifactId(ctx, "org.apache.commons", "commons-lang3")
//
//	// 设置每页大小并连接到客户端
//	iterator.SetPageSize(20)
//
//	// 使用迭代器处理所有结果
//	for iterator.HasNext() {
//	    artifacts, err := iterator.Next()
//	    if err != nil {
//	        log.Fatalf("获取下一批结果失败: %v", err)
//	    }
//
//	    for _, artifact := range artifacts {
//	        fmt.Printf("%s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	    }
//	}
func (c *Client) IteratorByGroupAndArtifactId(ctx context.Context, groupId, artifactId string) *SearchIterator[*response.Artifact] {
	query := request.NewQuery().SetGroupId(groupId).SetArtifactId(artifactId)
	search := request.NewSearchRequest().SetQuery(query)
	return NewSearchIterator[*response.Artifact](search).WithClient(c)
}

// GetArtifactDetails 获取制品的详细信息
//
// 该方法用于获取指定GroupId和ArtifactId的制品的详细信息，包括所有可用版本、制品的描述、许可证信息、
// 使用统计数据等元数据。它为开发者提供了对特定制品的全面了解，有助于评估制品的质量、流行度和适用性。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 要查询的制品的组ID，如"org.springframework"
//   - artifactId: 要查询的制品ID，如"spring-core"
//   - version: 制品版本，如为空则使用最新版本
//
// 返回:
//   - *response.ArtifactMetadata: 包含制品详细信息的对象
//   - error: 获取详情过程中发生的错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 获取Spring Framework Core的详细信息
//	details, err := client.GetArtifactDetails(ctx, "org.springframework", "spring-core", "")
//	if err != nil {
//	    log.Fatalf("获取制品详情失败: %v", err)
//	}
//
//	// 输出制品信息
//	fmt.Printf("制品: %s:%s\n", details.GroupId, details.ArtifactId)
//	fmt.Printf("最新版本: %s\n", details.LatestVersion)
//	fmt.Printf("发布时间: %s\n", details.LastUpdated)
//	fmt.Printf("打包类型: %s\n", details.Packaging)
//
//	// 查看依赖项
//	fmt.Printf("依赖项数量: %d\n", len(details.Dependencies))
//	for i, dep := range details.Dependencies {
//	    fmt.Printf("  %d. %s:%s:%s\n", i+1, dep.GroupId, dep.ArtifactId, dep.Version)
//	}
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
//
// 该方法提供了搜索当前流行或广泛使用的Maven制品的功能。搜索结果按照制品的流行度指标
// （如版本数量、使用统计等）降序排序，便于开发者快速发现和使用社区中最受欢迎的库和工具。
// 这对于寻找某类功能的最佳实现或评估不同制品的活跃度和社区支持情况非常有用。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - limit: 最大返回结果数量，控制返回的制品数
//
// 返回:
//   - []*response.Artifact: 热门制品列表，按流行度降序排序
//   - error: 搜索过程中发生的错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 搜索最热门的20个制品
//	artifacts, err := client.SearchPopularArtifacts(ctx, 20)
//	if err != nil {
//	    log.Fatalf("搜索热门制品失败: %v", err)
//	}
//
//	// 输出热门制品
//	fmt.Println("Maven仓库中最热门的制品:")
//	for i, artifact := range artifacts {
//	    fmt.Printf("%d. %s:%s:%s (下载量: %d)\n",
//	        i+1, artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion,
//	        artifact.DownloadCount)
//	}
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
//
// 该方法提供根据标签(Tag)搜索Maven制品的功能。标签通常用于对制品进行分类和归类，
// 帮助开发者快速找到特定领域或特定用途的制品。例如，可以搜索标记为"http-client"、
// "database"或"logging"等标签的所有制品，更有针对性地发现满足特定需求的库。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - tag: 要搜索的标签名称，如"json"、"http"、"database"等
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - []*response.Artifact: 具有指定标签的制品列表
//   - error: 搜索过程中发生的错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 搜索所有带"json"标签的制品
//	artifacts, err := client.SearchArtifactsByTag(ctx, "json", 20)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 输出搜索结果
//	fmt.Printf("找到 %d 个JSON相关的制品:\n", len(artifacts))
//	for i, artifact := range artifacts {
//	    fmt.Printf("%d. %s:%s:%s\n",
//	        i+1, artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
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
//
// 该方法提供了高级搜索功能，不仅返回匹配的制品列表，还返回聚合统计信息。
// 聚合功能可以帮助用户快速了解搜索结果在特定维度（如组ID、打包类型、许可证等）的分布情况。
// 这对于数据分析、趋势识别或结果过滤非常有用，可以帮助开发者更好地理解搜索结果的整体情况。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - searchText: 搜索文本，支持通配符和高级搜索语法
//   - facetFields: 要聚合的字段列表，如["g"(组ID), "p"(打包类型), "l"(许可证)]
//   - limit: 最大返回结果数量，控制返回的制品数
//
// 返回:
//   - []*response.Artifact: 匹配搜索条件的制品列表
//   - *response.FacetResults: 聚合结果，包含按指定字段分组的统计信息
//   - error: 搜索过程中发生的错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 搜索"logger"相关制品，并按组ID和打包类型聚合
//	artifacts, facets, err := client.SearchArtifactsWithFacets(ctx, "logger", []string{"g", "p"}, 50)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 输出搜索结果
//	fmt.Printf("找到 %d 个结果\n", len(artifacts))
//
//	// 分析组ID分布
//	fmt.Println("\n按组织分布:")
//	for _, facet := range facets.Counts["g"] {
//	    fmt.Printf("%s: %d 个制品\n", facet.Value, facet.Count)
//	}
//
//	// 分析打包类型分布
//	fmt.Println("\n按打包类型分布:")
//	for _, facet := range facets.Counts["p"] {
//	    fmt.Printf("%s: %d 个制品\n", facet.Value, facet.Count)
//	}
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
//
// 该方法用于获取指定Maven制品的完整依赖关系信息，包括直接依赖项、传递依赖项和可选依赖项。
// 依赖关系分析对于项目依赖管理、冲突检测、安全审计和兼容性评估至关重要。通过此方法，开发者
// 可以全面了解制品的依赖结构，有助于做出更明智的集成决策。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 要分析的制品的组ID，如"org.springframework"
//   - artifactId: 要分析的制品ID，如"spring-core"
//   - version: 要分析的制品版本，如"5.3.25"
//
// 返回:
//   - *ArtifactDependencyInfo: 包含分类依赖关系的对象
//   - error: 获取依赖关系过程中发生的错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 获取Spring Core 5.3.25的依赖关系
//	dependencies, err := client.GetArtifactDependencies(ctx, "org.springframework", "spring-core", "5.3.25")
//	if err != nil {
//	    log.Fatalf("获取依赖关系失败: %v", err)
//	}
//
//	// 输出直接依赖项
//	fmt.Printf("直接依赖项 (%d):\n", len(dependencies.DirectDependencies))
//	for i, dep := range dependencies.DirectDependencies {
//	    fmt.Printf("  %d. %s:%s:%s\n", i+1, dep.GroupId, dep.ArtifactId, dep.Version)
//	}
//
//	// 输出可选依赖项
//	fmt.Printf("\n可选依赖项 (%d):\n", len(dependencies.OptionalDependencies))
//	for i, dep := range dependencies.OptionalDependencies {
//	    fmt.Printf("  %d. %s:%s:%s\n", i+1, dep.GroupId, dep.ArtifactId, dep.Version)
//	}
//
//	// 输出传递依赖项
//	fmt.Printf("\n传递依赖项 (%d):\n", len(dependencies.TransitiveDependencies))
//	for i, dep := range dependencies.TransitiveDependencies {
//	    fmt.Printf("  %d. %s:%s:%s\n", i+1, dep.GroupId, dep.ArtifactId, dep.Version)
//	}
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
//
// 该方法用于分析指定Maven制品被其他项目使用的情况，包括总使用数量、主要使用者列表和按组分类的使用情况。
// 通过这些信息，开发者可以评估制品的流行度和影响力，了解主要的使用者是哪些项目或组织，有助于做出
// 关于API稳定性、向后兼容性和支持策略等方面的决策。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 要分析的制品的组ID，如"com.google.guava"
//   - artifactId: 要分析的制品ID，如"guava"
//   - version: 要分析的制品版本，如为空则分析所有版本
//   - limit: 返回的顶级使用者数量，控制结果列表大小
//
// 返回:
//   - *ArtifactUsage: 包含制品使用情况的详细信息
//   - error: 获取使用情况过程中发生的错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 获取Guava库的使用情况，返回前20个使用者
//	usage, err := client.GetArtifactUsage(ctx, "com.google.guava", "guava", "", 20)
//	if err != nil {
//	    log.Fatalf("获取使用情况失败: %v", err)
//	}
//
//	// 输出总使用情况
//	fmt.Printf("Guava库被%d个项目引用\n", usage.TotalUsageCount)
//
//	// 输出主要使用者
//	fmt.Println("\n主要使用者:")
//	for i, user := range usage.TopUsers {
//	    fmt.Printf("  %d. %s:%s:%s\n", i+1, user.GroupId, user.ArtifactId, user.LatestVersion)
//	}
//
//	// 输出按组分类的使用情况
//	fmt.Println("\n按组织分类的使用情况:")
//	for groupId, count := range usage.UsageByGroup {
//	    fmt.Printf("  %s: %d个项目\n", groupId, count)
//	}
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
//
// 该方法用于对比两个Maven制品的关键指标，包括版本数量、更新频率、活跃度和流行度等。
// 这种比较对于在多个相似库之间做选择时非常有用，可以帮助开发者评估哪个制品更活跃、
// 更受欢迎，或者哪个制品有更好的维护和社区支持。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId1: 第一个制品的组ID，如"org.apache.logging.log4j"
//   - artifactId1: 第一个制品的制品ID，如"log4j-core"
//   - groupId2: 第二个制品的组ID，如"ch.qos.logback"
//   - artifactId2: 第二个制品的制品ID，如"logback-classic"
//
// 返回:
//   - *ArtifactComparisonResult: 包含两个制品比较结果的详细信息
//   - error: 比较过程中发生的错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 比较两个日志框架：Log4j2和Logback
//	comparison, err := client.CompareArtifacts(ctx,
//	    "org.apache.logging.log4j", "log4j-core",
//	    "ch.qos.logback", "logback-classic")
//	if err != nil {
//	    log.Fatalf("比较制品失败: %v", err)
//	}
//
//	// 输出基本信息
//	fmt.Printf("比较 %s:%s 和 %s:%s\n",
//	    comparison.Artifact1.GroupId, comparison.Artifact1.ArtifactId,
//	    comparison.Artifact2.GroupId, comparison.Artifact2.ArtifactId)
//
//	// 输出版本信息
//	fmt.Printf("\n版本数量差异: %d\n", comparison.VersionCountDiff)
//	fmt.Printf("更新时间差异: %d天\n", comparison.UpdateTimeDiffDays)
//
//	// 输出活跃度和流行度比较
//	fmt.Printf("\n最活跃的制品: %s\n", comparison.MostActive)
//	fmt.Printf("最流行的制品: %s\n", comparison.MostPopular)
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
//
// 该方法提供了按发布或更新时间范围搜索Maven制品的功能。这对于发现特定时间段内
// 发布的新制品、监控活跃度变化趋势，或查找在特定时间点之后更新的制品特别有用。
// 结果默认按时间戳降序排序，便于快速查看最近更新的制品。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - startDate: 开始日期，格式为YYYY-MM-DD，如"2023-01-01"
//   - endDate: 结束日期，格式为YYYY-MM-DD，如"2023-12-31"
//   - limit: 最大返回结果数量，控制返回的制品数
//
// 返回:
//   - []*response.Artifact: 在指定日期范围内发布或更新的制品列表
//   - error: 搜索过程中发生的错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 搜索2023年发布的制品，限制返回50个结果
//	artifacts, err := client.SearchArtifactsByDateRange(ctx, "2023-01-01", "2023-12-31", 50)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 输出搜索结果
//	fmt.Printf("在2023年发布的制品 (%d个):\n", len(artifacts))
//	for i, artifact := range artifacts {
//	    // 将时间戳转换为可读格式
//	    updateTime := time.Unix(artifact.Timestamp/1000, 0).Format("2006-01-02")
//	    fmt.Printf("%d. %s:%s:%s (发布于: %s)\n",
//	        i+1, artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion, updateTime)
//	}
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
//
// 该方法提供基于给定制品的相似度推荐功能，可以帮助开发者发现功能相似或相关的其他Maven制品。
// 推荐算法基于标签匹配、文本相似度和制品描述等多个维度，尝试找出最相关的替代方案或补充工具。
// 这对于探索替代库、查找功能互补的工具或解决制品维护问题（如已弃用或不再维护的情况）特别有价值。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 基准制品的组ID，如"com.fasterxml.jackson.core"
//   - artifactId: 基准制品的制品ID，如"jackson-databind"
//   - limit: 最大返回结果数量，控制返回的推荐制品数
//
// 返回:
//   - []*response.Artifact: 与基准制品相似的推荐制品列表
//   - error: 推荐过程中发生的错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 查找与Jackson Databind相似的制品
//	similarArtifacts, err := client.SuggestSimilarArtifacts(ctx,
//	    "com.fasterxml.jackson.core", "jackson-databind", 10)
//	if err != nil {
//	    log.Fatalf("查找相似制品失败: %v", err)
//	}
//
//	// 输出推荐结果
//	fmt.Printf("与Jackson Databind相似的制品 (%d个):\n", len(similarArtifacts))
//	for i, artifact := range similarArtifacts {
//	    fmt.Printf("%d. %s:%s:%s\n",
//	        i+1, artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	    if len(artifact.Tags) > 0 {
//	        fmt.Printf("   标签: %s\n", strings.Join(artifact.Tags, ", "))
//	    }
//	}
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
//
// 该方法用于计算和返回指定Maven制品的综合统计指标，包括版本数量、更新历史、活跃度和流行度等。
// 这些统计数据对于评估制品的质量、稳定性、维护状况和社区支持程度非常有用，可以帮助开发者在
// 选择依赖项时做出更明智的决策。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 要分析的制品的组ID，如"org.apache.commons"
//   - artifactId: 要分析的制品ID，如"commons-lang3"
//
// 返回:
//   - *ArtifactStats: 包含制品统计信息的详细对象
//   - error: 获取统计信息过程中发生的错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 获取Apache Commons Lang3的统计信息
//	stats, err := client.GetArtifactStats(ctx, "org.apache.commons", "commons-lang3")
//	if err != nil {
//	    log.Fatalf("获取统计信息失败: %v", err)
//	}
//
//	// 输出基本信息
//	fmt.Printf("制品: %s:%s\n", stats.GroupId, stats.ArtifactId)
//	fmt.Printf("版本总数: %d\n", stats.TotalVersions)
//
//	// 输出时间信息
//	firstDate := time.Unix(stats.FirstVersionDate/1000, 0).Format("2006-01-02")
//	latestDate := time.Unix(stats.LatestVersionDate/1000, 0).Format("2006-01-02")
//	fmt.Printf("首个版本发布于: %s\n", firstDate)
//	fmt.Printf("最新版本发布于: %s\n", latestDate)
//	fmt.Printf("上次更新距今: %d天\n", stats.DaysSinceLastUpdate)
//
//	// 输出活跃度和流行度
//	fmt.Printf("更新频率: %.2f版本/月\n", stats.UpdateFrequency)
//	fmt.Printf("被引用次数: %d\n", stats.UsageCount)
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
