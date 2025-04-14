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
//
// 该方法用于在Maven Central仓库中搜索符合指定查询条件的制品，返回其GAV坐标信息。
// 查询使用Solr查询语法，可以组合多个条件，例如"g:org.apache AND a:commons*"。
// 查询结果会根据相关性排序，并限制返回数量。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - query: Solr格式的查询字符串，例如"g:org.apache AND a:commons*"，支持通配符和布尔运算符
//   - limit: 最大返回结果数量，建议值为10-100；如果为0，则使用服务器默认值(通常为10)
//
// 返回:
//   - []*response.Artifact: 符合条件的制品列表，每个制品包含GroupId、ArtifactId等元数据
//   - error: 如果请求失败或解析出错时返回相应错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 搜索Apache Commons库下的所有制品
//	artifacts, err := client.ListGAVs(ctx, "g:org.apache.commons", 20)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 打印搜索结果
//	for _, artifact := range artifacts {
//	    fmt.Printf("制品: %s:%s:%s\n",
//	        artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	    fmt.Printf("  - 最新更新: %s\n", artifact.Timestamp)
//	    fmt.Printf("  - 使用许可: %s\n", artifact.LicenseName)
//	}
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
//
// 该方法用于检索指定GroupId、ArtifactId和可选Version的制品详细信息。
// 它是精确查找特定制品的便捷方法，内部使用ListGAVs方法实现，但简化了查询参数构造过程。
// 如果未提供版本号，则返回最新版本的制品信息。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 制品的GroupId，如"org.apache.commons"
//   - artifactId: 制品的ArtifactId，如"commons-lang3"
//   - version: 制品的版本号，如"3.12.0"；如果为空字符串，则返回最新版本
//
// 返回:
//   - *response.Artifact: 制品的详细信息，包含元数据、许可证、最新版本等信息
//   - error: 如果请求失败、解析出错或制品不存在时返回相应错误；特别地，当制品不存在时返回ErrNotFound
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 获取特定版本的制品信息
//	artifact, err := client.GetGAVInfo(ctx, "com.google.guava", "guava", "31.1-jre")
//	if err != nil {
//	    if errors.Is(err, api.ErrNotFound) {
//	        fmt.Println("未找到指定的制品")
//	    } else {
//	        log.Fatalf("获取制品信息失败: %v", err)
//	    }
//	}
//
//	// 打印制品详情
//	fmt.Printf("制品: %s:%s:%s\n",
//	    artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	fmt.Printf("发布时间: %s\n", artifact.Timestamp)
//	fmt.Printf("许可证: %s\n", artifact.LicenseName)
//
//	// 获取最新版本的制品信息（不指定版本）
//	latestArtifact, err := client.GetGAVInfo(ctx, "org.apache.commons", "commons-lang3", "")
//	if err == nil {
//	    fmt.Printf("最新版本: %s\n", latestArtifact.LatestVersion)
//	}
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
//
// 该方法在Maven Central仓库中搜索制品，并支持按指定字段排序结果。
// 它是ListGAVs的高级版本，增加了排序功能。查询使用Solr查询语法，
// 可以通过sortField参数指定排序字段，如"timestamp"（发布时间）、
// "score"（相关度）、"artifactId"（按名称）等，并通过ascending参数
// 控制升序或降序排列。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - query: Solr格式的查询字符串，例如"g:org.springframework"
//   - sortField: 排序字段名称，常用的有timestamp、score、artifactId、groupId等
//   - ascending: 排序方向，true表示升序（从小到大），false表示降序（从大到小）
//   - limit: 最大返回结果数量，建议值为10-100；如果为0，则使用服务器默认值
//
// 返回:
//   - []*response.Artifact: 符合条件且已排序的制品列表
//   - error: 如果请求失败或解析出错时返回相应错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 查找Spring Framework相关制品，按发布时间降序排列（最新的排在前面）
//	artifacts, err := client.SearchGAVsWithSort(ctx, "g:org.springframework", "timestamp", false, 20)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 打印搜索结果
//	fmt.Println("Spring框架最近发布的制品:")
//	for i, artifact := range artifacts {
//	    fmt.Printf("%d. %s:%s:%s (发布于: %s)\n",
//	        i+1, artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion, artifact.Timestamp)
//	}
//
//	// 查找Apache Commons库下的所有制品，按名称字母顺序排列
//	alphabeticalArtifacts, err := client.SearchGAVsWithSort(ctx, "g:org.apache.commons", "artifactId", true, 20)
//	if err != nil {
//	    log.Fatalf("搜索失败: %v", err)
//	}
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
//
// 该方法尝试查找在Maven Central仓库中两个制品之间的依赖关系。
// 目前这个方法主要返回第一个制品的信息，完整的依赖分析需要进一步解析POM文件。
// 这是一个实验性功能，因为Maven Central API并不直接提供依赖关系查询，
// 需要下载POM文件并解析其中的依赖声明。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId1: 源制品的GroupId
//   - artifactId1: 源制品的ArtifactId
//   - groupId2: 目标制品的GroupId，用于后续依赖分析（当前版本未完全实现）
//   - artifactId2: 目标制品的ArtifactId，用于后续依赖分析（当前版本未完全实现）
//   - limit: 最大返回结果数量
//
// 返回:
//   - []*response.Artifact: 源制品的信息列表，后续版本会增加依赖关系分析
//   - error: 如果请求失败或解析出错时返回相应错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 查找commons-lang3可能依赖的jackson-core制品
//	artifacts, err := client.FindGAVDependencies(
//	    ctx,
//	    "org.apache.commons", "commons-lang3",
//	    "com.fasterxml.jackson.core", "jackson-core",
//	    10)
//	if err != nil {
//	    log.Fatalf("查询依赖关系失败: %v", err)
//	}
//
//	// 打印找到的制品(注意:完整的依赖分析需要进一步处理)
//	for _, artifact := range artifacts {
//	    fmt.Printf("源制品: %s:%s:%s\n",
//	        artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	    // 要确定实际依赖关系，需要进一步下载和解析POM文件
//	}
//
//	// 要完成完整的依赖分析，可能需要：
//	// 1. 下载源制品的POM文件
//	// 2. 解析POM文件中的依赖声明
//	// 3. 检查是否包含目标制品的依赖
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
//
// 该方法提供了分页功能的GAV查询，适用于需要处理大量搜索结果的场景。
// 通过指定页码和每页大小，可以分批次获取结果，避免一次性加载过多数据导致的内存压力。
// 除了返回当前页的制品列表外，还会返回符合条件的总记录数，便于实现分页导航。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - query: Solr格式的查询字符串，例如"g:com.google.guava"
//   - page: 页码，从1开始；如果小于1，将默认使用第1页
//   - pageSize: 每页记录数；如果小于1，将默认使用10条/页
//
// 返回:
//   - []*response.Artifact: 当前页的制品列表
//   - int: 符合查询条件的总记录数，用于计算总页数
//   - error: 如果请求失败或解析出错时返回相应错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 分页查询Apache Commons库下的所有制品
//	query := "g:org.apache.commons"
//	pageSize := 20
//
//	// 获取第一页数据
//	artifacts, total, err := client.ListGAVsPaginated(ctx, query, 1, pageSize)
//	if err != nil {
//	    log.Fatalf("查询失败: %v", err)
//	}
//
//	// 计算总页数并显示分页信息
//	totalPages := (total + pageSize - 1) / pageSize
//	fmt.Printf("查询结果: 共找到 %d 个制品，分 %d 页显示，当前第 %d 页\n",
//	    total, totalPages, 1)
//
//	// 显示当前页的数据
//	for i, artifact := range artifacts {
//	    fmt.Printf("%d. %s:%s:%s\n",
//	        i+1, artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
//
//	// 如果需要，可以获取下一页数据
//	if totalPages > 1 {
//	    nextPageArtifacts, _, err := client.ListGAVsPaginated(ctx, query, 2, pageSize)
//	    // 处理下一页数据...
//	}
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
//
// 该方法创建一个迭代器对象，用于高效地处理可能包含大量结果的GAV查询。
// 迭代器通过"惰性加载"方式工作，只有在需要时才会发起网络请求获取下一批数据，
// 这种方式特别适合处理结果数量未知或非常大的查询，可以有效控制内存使用并提供
// 流式处理能力。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - query: Solr格式的查询字符串，例如"g:org.apache*"
//
// 返回:
//   - *SearchIterator[*response.Artifact]: 可用于遍历搜索结果的迭代器对象
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建一个迭代器，查询所有Apache项目
//	iterator := client.IteratorGAVs(ctx, "g:org.apache*")
//
//	// 使用迭代器遍历所有结果
//	totalProcessed := 0
//	for iterator.HasNext() {
//	    // 获取下一批结果(默认每批10条)
//	    artifacts, err := iterator.Next()
//	    if err != nil {
//	        log.Fatalf("获取下一批数据失败: %v", err)
//	    }
//
//	    // 处理这一批数据
//	    for _, artifact := range artifacts {
//	        totalProcessed++
//	        // 示例：仅打印前100个结果
//	        if totalProcessed <= 100 {
//	            fmt.Printf("%d. %s:%s:%s\n",
//	                totalProcessed, artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	        }
//	    }
//
//	    // 可以随时中断处理
//	    if totalProcessed > 1000 {
//	        fmt.Println("已处理超过1000条记录，停止遍历")
//	        break
//	    }
//	}
//
//	fmt.Printf("总共处理了 %d 条记录\n", totalProcessed)
//
//	// 或者直接将所有结果转换为切片(注意:如果结果很多，这可能会消耗大量内存)
//	// allArtifacts, err := iterator.ToSlice()
func (c *Client) IteratorGAVs(ctx context.Context, query string) *SearchIterator[*response.Artifact] {
	searchRequest := request.NewSearchRequest()
	searchRequest.Query.SetCustomQuery(query)
	searchRequest.SetCore("gav")
	searchRequest.AddCustomParam("wt", "json")

	return NewSearchIterator[*response.Artifact](searchRequest).WithClient(c)
}
