package api

import (
	"context"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// AsyncSearchByGroup 异步搜索组ID
//
// 该方法提供根据组ID进行异步搜索的功能，允许应用程序在不阻塞主线程的情况下执行搜索。
// 内部使用goroutine执行实际的搜索操作，并通过通道返回结果，适合需要同时执行多个搜索
// 或在UI线程中保持响应性的场景。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 要搜索的组ID，如"org.apache.commons"
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - <-chan AsyncResult[[]*response.Artifact]: 用于接收异步搜索结果的只读通道
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 启动异步搜索
//	resultChan := client.AsyncSearchByGroup(ctx, "org.apache.commons", 10)
//
//	// 继续执行其他操作...
//
//	// 稍后获取结果
//	result := <-resultChan
//	if result.Error != nil {
//	    log.Fatalf("搜索失败: %v", result.Error)
//	}
//
//	// 处理搜索结果
//	artifacts := result.Result
//	fmt.Printf("找到 %d 个结果\n", len(artifacts))
//	for _, artifact := range artifacts {
//	    fmt.Printf("%s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
func (c *Client) AsyncSearchByGroup(ctx context.Context, groupId string, limit int) <-chan AsyncResult[[]*response.Artifact] {
	resultChan := make(chan AsyncResult[[]*response.Artifact], 1)

	go func() {
		defer close(resultChan)

		result, err := c.SearchByGroupId(ctx, groupId, limit)
		resultChan <- AsyncResult[[]*response.Artifact]{
			Result:  result,
			Error:   err,
			Context: groupId,
		}
	}()

	return resultChan
}

// AsyncSearchByArtifact 异步搜索制品ID
//
// 该方法提供根据制品ID进行异步搜索的功能，在后台线程中执行搜索并通过通道返回结果。
// 这种异步模式适合需要同时执行多个搜索操作或在前台保持UI响应性的场景。
// 每个搜索操作都会在单独的goroutine中执行，避免阻塞主线程。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - artifactId: 要搜索的制品ID，如"commons-lang3"
//   - limit: 最大返回结果数量，如果小于等于0则返回所有结果
//
// 返回:
//   - <-chan AsyncResult[[]*response.Artifact]: 用于接收异步搜索结果的只读通道
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 启动异步搜索
//	resultChan := client.AsyncSearchByArtifact(ctx, "guava", 5)
//
//	// 继续执行其他操作...
//
//	// 在适当的时候获取结果
//	result := <-resultChan
//	if result.Error != nil {
//	    log.Fatalf("搜索失败: %v", result.Error)
//	}
//
//	// 处理搜索结果
//	artifacts := result.Result
//	fmt.Printf("找到 %d 个'guava'制品\n", len(artifacts))
//	for _, artifact := range artifacts {
//	    fmt.Printf("%s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
func (c *Client) AsyncSearchByArtifact(ctx context.Context, artifactId string, limit int) <-chan AsyncResult[[]*response.Artifact] {
	resultChan := make(chan AsyncResult[[]*response.Artifact], 1)

	go func() {
		defer close(resultChan)

		result, err := c.SearchByArtifactId(ctx, artifactId, limit)
		resultChan <- AsyncResult[[]*response.Artifact]{
			Result:  result,
			Error:   err,
			Context: artifactId,
		}
	}()

	return resultChan
}

// AsyncBatchSearch 批量异步搜索
//
// 该方法允许同时执行多个不同的搜索请求，每个请求在独立的goroutine中并行处理。
// 这种设计特别适合需要同时查询多种不同条件的场景，显著提高搜索效率。
// 所有搜索结果将通过同一个通道返回，每个结果包含对应的请求上下文以便识别。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - requests: 搜索请求对象数组，每个对象可以包含不同的搜索条件
//
// 返回:
//   - <-chan AsyncResult[[]*response.Artifact]: 用于接收所有异步搜索结果的只读通道
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建多个不同的搜索请求
//	request1 := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId("org.apache.commons"))
//	request2 := request.NewSearchRequest().SetQuery(request.NewQuery().SetArtifactId("junit"))
//	request3 := request.NewSearchRequest().SetQuery(request.NewQuery().SetClassName("Logger"))
//
//	// 执行批量异步搜索
//	resultChan := client.AsyncBatchSearch(ctx, []*request.SearchRequest{request1, request2, request3})
//
//	// 处理所有搜索结果
//	var results1, results2, results3 []*response.Artifact
//	for i := 0; i < 3; i++ {
//	    result := <-resultChan
//	    if result.Error != nil {
//	        log.Printf("搜索失败: %v", result.Error)
//	        continue
//	    }
//
//	    // 根据上下文区分结果
//	    switch req := result.Context.(*request.SearchRequest); {
//	    case req == request1:
//	        results1 = result.Result
//	    case req == request2:
//	        results2 = result.Result
//	    case req == request3:
//	        results3 = result.Result
//	    }
//	}
//
//	// 使用结果
//	fmt.Printf("Apache Commons库: %d 个结果\n", len(results1))
//	fmt.Printf("JUnit相关制品: %d 个结果\n", len(results2))
//	fmt.Printf("Logger类: %d 个结果\n", len(results3))
func (c *Client) AsyncBatchSearch(ctx context.Context, requests []*request.SearchRequest) <-chan AsyncResult[[]*response.Artifact] {
	resultChan := make(chan AsyncResult[[]*response.Artifact], len(requests))

	// 启动所有请求
	for _, req := range requests {
		go func(searchReq *request.SearchRequest) {
			result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, searchReq)
			if err != nil {
				resultChan <- AsyncResult[[]*response.Artifact]{
					Error:   err,
					Context: searchReq,
				}
				return
			}

			resultChan <- AsyncResult[[]*response.Artifact]{
				Result:  result.ResponseBody.Docs,
				Context: searchReq,
			}
		}(req)
	}

	return resultChan
}

// AsyncGetArtifactMetadata 异步获取制品元数据
//
// 该方法提供了异步获取Maven制品元数据的功能，通过后台goroutine执行实际的API请求，
// 避免阻塞调用线程。这种异步模式特别适合在需要获取多个制品元数据或在UI线程中
// 保持响应性的场景中使用。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 制品的组ID，如"org.apache.commons"
//   - artifactId: 制品的ID，如"commons-lang3"
//   - version: 制品的版本号，如"3.12.0"
//
// 返回:
//   - <-chan AsyncResult[*response.ArtifactMetadata]: 用于接收异步元数据结果的只读通道
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 启动异步元数据获取
//	resultChan := client.AsyncGetArtifactMetadata(ctx, "com.google.guava", "guava", "31.1-jre")
//
//	// 继续执行其他操作...
//
//	// 稍后获取结果
//	result := <-resultChan
//	if result.Error != nil {
//	    log.Fatalf("获取元数据失败: %v", result.Error)
//	}
//
//	// 使用元数据
//	metadata := result.Result
//	fmt.Printf("制品: %s:%s:%s\n", metadata.GroupId, metadata.ArtifactId, metadata.Version)
//	fmt.Printf("最后更新: %s\n", metadata.LastUpdated)
//	fmt.Printf("依赖项数量: %d\n", len(metadata.Dependencies))
//
//	// 可以通过Context获取原始请求的GAV坐标
//	if ctx, ok := result.Context.(map[string]string); ok {
//	    fmt.Printf("请求的GAV: %s:%s:%s\n", ctx["groupId"], ctx["artifactId"], ctx["version"])
//	}
func (c *Client) AsyncGetArtifactMetadata(ctx context.Context, groupId, artifactId, version string) <-chan AsyncResult[*response.ArtifactMetadata] {
	resultChan := make(chan AsyncResult[*response.ArtifactMetadata], 1)

	go func() {
		defer close(resultChan)

		result, err := c.GetArtifactMetadata(ctx, groupId, artifactId, version)
		resultChan <- AsyncResult[*response.ArtifactMetadata]{
			Result: result,
			Error:  err,
			Context: map[string]string{
				"groupId":    groupId,
				"artifactId": artifactId,
				"version":    version,
			},
		}
	}()

	return resultChan
}
