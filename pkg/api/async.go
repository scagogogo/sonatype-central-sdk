package api

import (
	"context"
	"sync"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// AsyncResult 异步操作结果
type AsyncResult[T any] struct {
	// 操作结果
	Result T

	// 错误信息
	Error error

	// 上下文信息
	Context interface{}
}

// AsyncSearchRequest 异步执行搜索请求并返回结果通道
//
// 该函数提供搜索请求的异步执行机制，允许应用程序在不阻塞主线程的情况下执行搜索操作。
// 函数会在后台线程中执行搜索，并通过通道返回结果。这种模式特别适合需要并行处理多个
// 搜索请求，或在执行长时间搜索时保持UI响应性的场景。
//
// 参数:
//   - c: API客户端实例，包含基础URL和HTTP配置
//   - ctx: 请求上下文，用于控制超时和取消
//   - searchRequest: 包含查询参数、分页、排序等信息的搜索请求对象
//
// 返回:
//   - <-chan AsyncResult[*response.Response[Doc]]: 用于接收异步搜索结果的只读通道
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建搜索请求
//	query := request.NewQuery().SetGroupId("org.apache.commons")
//	searchReq := request.NewSearchRequest().SetQuery(query).SetLimit(10)
//
//	// 异步执行搜索
//	resultChan := api.AsyncSearchRequest[*response.Artifact](client, ctx, searchReq)
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
//	response := result.Result
//	fmt.Printf("找到 %d 个结果\n", response.ResponseBody.NumFound)
//	for _, artifact := range response.ResponseBody.Docs {
//	    fmt.Printf("%s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
func AsyncSearchRequest[Doc any](c *Client, ctx context.Context, searchRequest *request.SearchRequest) <-chan AsyncResult[*response.Response[Doc]] {
	// 创建带缓冲的通道，避免发送方在接收方尚未准备好时阻塞
	resultChan := make(chan AsyncResult[*response.Response[Doc]], 1)

	// 在单独的goroutine中执行搜索操作
	go func() {
		// 确保在返回结果后关闭通道，防止接收方无限等待
		defer close(resultChan)

		// 执行实际的搜索请求
		result, err := SearchRequestJsonDoc[Doc](c, ctx, searchRequest)

		// 将结果和可能的错误发送到通道
		resultChan <- AsyncResult[*response.Response[Doc]]{
			Result:  result,
			Error:   err,
			Context: searchRequest, // 包含原始请求以供参考
		}
	}()

	return resultChan
}

// AsyncSearchRequestDoc 异步执行搜索请求并返回结果通道（别名函数）
//
// 该函数是AsyncSearchRequest的别名，提供完全相同的功能，保留此函数是为了
// 向后兼容。新代码应优先使用AsyncSearchRequest函数。
//
// 参数:
//   - c: API客户端实例，包含基础URL和HTTP配置
//   - ctx: 请求上下文，用于控制超时和取消
//   - searchRequest: 包含查询参数、分页、排序等信息的搜索请求对象
//
// 返回:
//   - <-chan AsyncResult[*response.Response[Doc]]: 用于接收异步搜索结果的只读通道
//
// 使用示例: 参见AsyncSearchRequest函数的示例
func AsyncSearchRequestDoc[Doc any](c *Client, ctx context.Context, searchRequest *request.SearchRequest) <-chan AsyncResult[*response.Response[Doc]] {
	// 直接调用AsyncSearchRequest，保持实现一致性
	return AsyncSearchRequest[Doc](c, ctx, searchRequest)
}

// BatchAsyncSearch 批量异步搜索
//
// 该方法提供高效的批量异步搜索功能，允许同时执行多个搜索请求。
// 每个请求在独立的goroutine中并行处理，通过单一通道返回所有结果。
// 设计用于需要执行多个不同搜索条件的场景，显著提高搜索效率。
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
//	resultChan := client.BatchAsyncSearch(ctx, []*request.SearchRequest{request1, request2, request3})
//
//	// 处理所有搜索结果
//	var results1, results2, results3 []*response.Artifact
//	resultCount := 0
//	for result := range resultChan {
//	    resultCount++
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
//
//	    // 所有结果都处理完后退出
//	    if resultCount == 3 {
//	        break
//	    }
//	}
//
//	// 使用结果
//	fmt.Printf("Apache Commons库: %d 个结果\n", len(results1))
//	fmt.Printf("JUnit相关制品: %d 个结果\n", len(results2))
//	fmt.Printf("Logger类: %d 个结果\n", len(results3))
func (c *Client) BatchAsyncSearch(ctx context.Context, requests []*request.SearchRequest) <-chan AsyncResult[[]*response.Artifact] {
	resultChan := make(chan AsyncResult[[]*response.Artifact], len(requests))

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

	// 创建结果收集器
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(len(requests))

		// 等待所有请求完成
		go func() {
			wg.Wait()
			close(resultChan)
		}()
	}()

	return resultChan
}

// AsyncSearchByGroupId 异步搜索GroupId
//
// 该方法提供根据GroupId进行异步搜索的功能，允许应用程序在不阻塞主线程的情况下执行搜索。
// 内部使用goroutine执行实际的搜索操作，并通过通道返回结果，适合需要同时执行多个搜索
// 或在UI应用程序中防止搜索操作阻塞界面响应的场景。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 要搜索的Maven坐标groupId
//   - limit: 结果数量限制，控制返回的最大制品数
//
// 返回:
//   - <-chan AsyncResult[[]*response.Artifact]: 用于接收异步搜索结果的只读通道
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 异步搜索Apache Commons库
//	resultChan := client.AsyncSearchByGroupId(ctx, "org.apache.commons", 20)
//
//	// 可以在等待结果期间执行其他操作
//	// ...
//
//	// 获取搜索结果
//	result := <-resultChan
//	if result.Error != nil {
//	    log.Fatalf("搜索失败: %v", result.Error)
//	}
//
//	// 处理搜索结果
//	artifacts := result.Result
//	fmt.Printf("找到 %d 个Apache Commons库:\n", len(artifacts))
//	for _, artifact := range artifacts {
//	    fmt.Printf("%s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
func (c *Client) AsyncSearchByGroupId(ctx context.Context, groupId string, limit int) <-chan AsyncResult[[]*response.Artifact] {
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

// AsyncSearchByArtifactId 异步搜索ArtifactId
//
// 该方法提供根据ArtifactId进行异步搜索的功能，允许应用程序在不阻塞主线程的情况下执行搜索。
// 内部使用goroutine执行实际的搜索操作，并通过通道返回结果，适合需要同时执行多个搜索
// 或在UI应用程序中防止搜索操作阻塞界面响应的场景。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - artifactId: 要搜索的Maven坐标artifactId
//   - limit: 结果数量限制，控制返回的最大制品数
//
// 返回:
//   - <-chan AsyncResult[[]*response.Artifact]: 用于接收异步搜索结果的只读通道
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 异步搜索所有名称为"guava"的制品
//	resultChan := client.AsyncSearchByArtifactId(ctx, "guava", 10)
//
//	// 可以在等待结果期间执行其他操作
//	// ...
//
//	// 获取搜索结果
//	result := <-resultChan
//	if result.Error != nil {
//	    log.Fatalf("搜索失败: %v", result.Error)
//	}
//
//	// 处理搜索结果
//	artifacts := result.Result
//	fmt.Printf("找到 %d 个'guava'制品:\n", len(artifacts))
//	for _, artifact := range artifacts {
//	    fmt.Printf("%s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//	}
func (c *Client) AsyncSearchByArtifactId(ctx context.Context, artifactId string, limit int) <-chan AsyncResult[[]*response.Artifact] {
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

// AsyncDownload 异步下载文件
//
// 该方法提供异步下载Maven仓库中文件的功能，在后台线程中执行下载操作并通过通道返回结果。
// 这种异步设计使应用程序能够在下载大文件期间保持响应性，或并行下载多个文件。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - filePath: 要下载的文件路径，相对于Maven仓库基础URL
//
// 返回:
//   - <-chan AsyncResult[[]byte]: 用于接收异步下载结果的只读通道，结果为文件内容的字节数组
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 启动异步下载
//	jarPath := "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar"
//	resultChan := client.AsyncDownload(ctx, jarPath)
//
//	// 继续执行其他操作...
//
//	// 稍后获取下载结果
//	result := <-resultChan
//	if result.Error != nil {
//	    log.Fatalf("下载失败: %v", result.Error)
//	}
//
//	// 使用下载的文件内容
//	fileBytes := result.Result
//	fmt.Printf("已下载 %s，大小: %d 字节\n", jarPath, len(fileBytes))
//
//	// 将内容保存到本地文件
//	err := os.WriteFile("commons-lang3.jar", fileBytes, 0644)
//	if err != nil {
//	    log.Fatalf("保存文件失败: %v", err)
//	}
func (c *Client) AsyncDownload(ctx context.Context, filePath string) <-chan AsyncResult[[]byte] {
	resultChan := make(chan AsyncResult[[]byte], 1)

	go func() {
		defer close(resultChan)

		result, err := c.Download(ctx, filePath)
		resultChan <- AsyncResult[[]byte]{
			Result:  result,
			Error:   err,
			Context: filePath,
		}
	}()

	return resultChan
}

// AsyncBatchDownload 异步批量下载文件
//
// 该方法提供高效的批量异步下载功能，允许同时下载多个文件。每个文件在独立的goroutine中
// 并行下载，所有下载结果通过单一通道返回。这种设计特别适合需要同时获取多个相关文件的场景，
// 如下载Maven制品的jar、源码和文档。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - filePaths: 要下载的文件路径数组，每个路径相对于Maven仓库基础URL
//
// 返回:
//   - <-chan AsyncResult[[]byte]: 用于接收所有异步下载结果的只读通道
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 准备要下载的文件路径列表
//	basePath := "com/google/guava/guava/31.1-jre"
//	filePaths := []string{
//	    basePath + "/guava-31.1-jre.jar",
//	    basePath + "/guava-31.1-jre-sources.jar",
//	    basePath + "/guava-31.1-jre-javadoc.jar",
//	}
//
//	// 启动批量异步下载
//	resultChan := client.AsyncBatchDownload(ctx, filePaths)
//
//	// 处理所有下载结果
//	downloadResults := make(map[string][]byte)
//	failedDownloads := make(map[string]error)
//
//	for result := range resultChan {
//	    filePath := result.Context.(string)
//	    if result.Error != nil {
//	        failedDownloads[filePath] = result.Error
//	        continue
//	    }
//	    downloadResults[filePath] = result.Result
//	}
//
//	// 使用下载结果
//	for path, content := range downloadResults {
//	    fileName := filepath.Base(path)
//	    fmt.Printf("已下载 %s，大小: %d 字节\n", fileName, len(content))
//
//	    // 保存到本地文件系统
//	    err := os.WriteFile(fileName, content, 0644)
//	    if err != nil {
//	        log.Printf("保存文件 %s 失败: %v", fileName, err)
//	    }
//	}
//
//	// 报告失败的下载
//	for path, err := range failedDownloads {
//	    fmt.Printf("下载 %s 失败: %v\n", path, err)
//	}
func (c *Client) AsyncBatchDownload(ctx context.Context, filePaths []string) <-chan AsyncResult[[]byte] {
	resultChan := make(chan AsyncResult[[]byte], len(filePaths))

	for _, path := range filePaths {
		go func(filePath string) {
			result, err := c.Download(ctx, filePath)
			resultChan <- AsyncResult[[]byte]{
				Result:  result,
				Error:   err,
				Context: filePath,
			}
		}(path)
	}

	// 创建结果收集器
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(len(filePaths))

		// 等待所有请求完成
		go func() {
			wg.Wait()
			close(resultChan)
		}()
	}()

	return resultChan
}
