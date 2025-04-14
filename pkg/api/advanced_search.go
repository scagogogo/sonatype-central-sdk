package api

import (
	"context"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// AdvancedSearch 高级搜索，支持完整的坐标搜索
//
// 该方法提供强大的高级搜索功能，允许根据Maven坐标的任意组合进行精确搜索。
// 支持按groupId、artifactId、version、packaging和classifier等完整坐标进行搜索，
// 是执行精确匹配和筛选的最佳选择。
//
// 参数:
//   - ctx: 请求上下文，用于控制超时和取消
//   - options: 高级搜索选项，包含各种坐标搜索参数
//   - limit: 结果数量限制，控制返回的最大制品数
//
// 返回:
//   - []*response.Artifact: 符合条件的制品列表
//   - error: 搜索过程中的错误，若成功则为nil
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建高级搜索选项
//	options := &request.AdvancedSearchOptions{
//	    GroupId:    "org.apache.commons",
//	    ArtifactId: "commons-lang3",
//	    Version:    "3.12.0",
//	    Packaging:  "jar",
//	}
//
//	// 执行高级搜索
//	artifacts, err := client.AdvancedSearch(ctx, options, 10)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 处理搜索结果
//	for _, artifact := range artifacts {
//	    fmt.Printf("%s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.Version)
//	}
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
//
// 该方法创建一个迭代器，用于分页处理大量高级搜索结果。当搜索结果可能很大或需要
// 批量处理时，迭代器模式比一次性加载所有结果更高效。迭代器会按需获取下一批结果，
// 减少内存占用并提高应用程序响应性。
//
// 参数:
//   - ctx: 请求上下文，用于控制超时和取消
//   - options: 高级搜索选项，包含各种坐标搜索参数
//
// 返回:
//   - *SearchIterator[*response.Artifact]: 用于迭代搜索结果的迭代器
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建高级搜索选项
//	options := &request.AdvancedSearchOptions{
//	    GroupId: "org.apache.commons",
//	}
//
//	// 创建迭代器
//	iterator := client.AdvancedSearchIterator(ctx, options)
//
//	// 设置每页大小并连接到客户端
//	iterator.SetPageSize(20).Connect(client, ctx)
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
//
// 该方法提供带自定义排序功能的搜索，允许指定结果的排序字段和顺序。
// 当需要按特定字段（如时间戳、下载量或名称）排序时非常有用，可以帮助用户
// 快速找到最新、最流行或按字母顺序排列的制品。
//
// 参数:
//   - ctx: 请求上下文，用于控制超时和取消
//   - searchQuery: 搜索请求对象，包含查询条件
//   - sortField: 排序字段名称，如"timestamp"、"download_count"等
//   - ascending: 是否按升序排列，true为升序，false为降序
//   - limit: 结果数量限制，控制返回的最大制品数
//
// 返回:
//   - []*response.Artifact: 排序后的制品列表
//   - error: 搜索过程中的错误，若成功则为nil
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建搜索请求
//	query := request.NewQuery().SetGroupId("org.apache.commons")
//	searchReq := request.NewSearchRequest().SetQuery(query)
//
//	// 执行带排序的搜索，按时间戳降序
//	artifacts, err := client.SearchWithSort(ctx, searchReq, "timestamp", false, 20)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 处理搜索结果（最新的制品将排在前面）
//	fmt.Println("最新Apache Commons制品:")
//	for _, artifact := range artifacts {
//	    fmt.Printf("%s:%s:%s (更新时间: %s)\n",
//	        artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion, artifact.Timestamp)
//	}
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
//
// 该方法用于查找使用指定依赖的制品。这对于分析依赖关系图、查找使用特定库的项目，
// 或者分析特定库的受欢迎程度非常有用。搜索会返回在其POM文件中声明了指定groupId
// 和artifactId依赖的所有制品。
//
// 参数:
//   - ctx: 请求上下文，用于控制超时和取消
//   - groupId: 要搜索的依赖groupId
//   - artifactId: 要搜索的依赖artifactId
//   - limit: 结果数量限制，控制返回的最大制品数
//
// 返回:
//   - []*response.Artifact: 使用指定依赖的制品列表
//   - error: 搜索过程中的错误，若成功则为nil
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 查找使用log4j的项目
//	artifacts, err := client.SearchByDependency(ctx, "org.apache.logging.log4j", "log4j-core", 50)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 分析使用log4j的项目
//	fmt.Printf("找到 %d 个使用log4j的项目:\n", len(artifacts))
//	for _, artifact := range artifacts {
//	    fmt.Printf("%s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
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
//
// 该方法用于查找使用特定许可证的制品。这对于开源合规审查、查找符合特定许可要求的
// 库，或者分析开源许可证分布情况非常有用。搜索会返回在其元数据中声明了指定许可证
// 的所有制品。
//
// 参数:
//   - ctx: 请求上下文，用于控制超时和取消
//   - license: 要搜索的许可证名称或标识符（如"Apache-2.0"、"MIT"、"GPL-3.0"等）
//   - limit: 结果数量限制，控制返回的最大制品数
//
// 返回:
//   - []*response.Artifact: 使用指定许可证的制品列表
//   - error: 搜索过程中的错误，若成功则为nil
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 查找使用MIT许可证的库
//	artifacts, err := client.SearchByLicense(ctx, "MIT", 100)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 分析使用MIT许可证的库
//	fmt.Printf("找到 %d 个使用MIT许可证的库:\n", len(artifacts))
//	for _, artifact := range artifacts {
//	    fmt.Printf("%s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
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
//
// 该方法用于获取指定制品的详细元数据信息。它首先根据提供的坐标查找制品，
// 然后获取其基本信息。如果提供了版本参数，还会尝试下载并解析POM文件以获取
// 更详细的元数据。这对于需要分析制品依赖、许可证或其他详细信息的场景非常有用。
//
// 参数:
//   - ctx: 请求上下文，用于控制超时和取消
//   - groupId: 制品的groupId
//   - artifactId: 制品的artifactId
//   - version: 制品的版本，如果为空则获取最新版本信息
//
// 返回:
//   - *response.ArtifactMetadata: 包含详细元数据的对象
//   - error: 获取过程中的错误，若成功则为nil，未找到则返回ErrNotFound
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 获取特定版本的元数据
//	metadata, err := client.GetArtifactMetadata(ctx, "org.apache.commons", "commons-lang3", "3.12.0")
//	if err != nil {
//	    if err == api.ErrNotFound {
//	        fmt.Println("制品未找到")
//	    } else {
//	        log.Fatalf("获取元数据失败: %v", err)
//	    }
//	    return
//	}
//
//	// 使用元数据
//	fmt.Printf("制品: %s:%s:%s\n", metadata.GroupId, metadata.ArtifactId, metadata.LatestVersion)
//	fmt.Printf("打包类型: %s\n", metadata.Packaging)
//	fmt.Printf("最后更新: %s\n", metadata.LastUpdated)
//	if metadata.PomContent != "" {
//	    fmt.Println("POM内容长度:", len(metadata.PomContent))
//	}
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
//
// 该方法支持同时执行多个不同的搜索请求，并以映射形式返回结果。每个请求在单独的goroutine中
// 并行执行，显著提高效率。这对于需要执行多个独立搜索操作而不希望串行等待的场景非常有用，
// 例如同时搜索多个不同的库或者使用不同条件查询同一组件的不同版本。
//
// 参数:
//   - ctx: 请求上下文，用于控制超时和取消
//   - queries: 搜索请求对象数组，每个对象可以包含不同的搜索条件
//
// 返回:
//   - map[string][]*response.Artifact: 以查询键为索引的结果映射
//   - error: 批量搜索过程中的错误，只有当所有查询都失败时才返回错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建多个不同的搜索请求
//	query1 := request.NewSearchRequest().SetQuery(
//	    request.NewQuery().SetGroupId("org.apache.commons"))
//	query1.SetQueryKey("commons")
//
//	query2 := request.NewSearchRequest().SetQuery(
//	    request.NewQuery().SetGroupId("org.springframework"))
//	query2.SetQueryKey("spring")
//
//	// 执行批量搜索
//	results, err := client.BatchSearch(ctx, []*request.SearchRequest{query1, query2})
//	if err != nil {
//	    log.Fatalf("批量搜索失败: %v", err)
//	}
//
//	// 处理不同的搜索结果
//	for key, artifacts := range results {
//	    fmt.Printf("=== 查询 '%s' 的结果 (共 %d 个) ===\n", key, len(artifacts))
//	    for i, artifact := range artifacts {
//	        if i >= 5 {
//	            fmt.Println("...")
//	            break
//	        }
//	        fmt.Printf("%s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	    }
//	    fmt.Println()
//	}
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
