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

// AsyncSearchRequest 异步执行搜索请求
// 注意：由于Go不支持泛型方法，这里改为类型参数化的函数
func AsyncSearchRequest[Doc any](c *Client, ctx context.Context, searchRequest *request.SearchRequest) <-chan AsyncResult[*response.Response[Doc]] {
	resultChan := make(chan AsyncResult[*response.Response[Doc]], 1)

	go func() {
		defer close(resultChan)

		result, err := SearchRequestJsonDoc[Doc](c, ctx, searchRequest)
		resultChan <- AsyncResult[*response.Response[Doc]]{
			Result:  result,
			Error:   err,
			Context: searchRequest,
		}
	}()

	return resultChan
}

// AsyncSearchRequestDoc 异步执行搜索请求（函数版本）
func AsyncSearchRequestDoc[Doc any](c *Client, ctx context.Context, searchRequest *request.SearchRequest) <-chan AsyncResult[*response.Response[Doc]] {
	resultChan := make(chan AsyncResult[*response.Response[Doc]], 1)

	go func() {
		defer close(resultChan)

		result, err := SearchRequestJsonDoc[Doc](c, ctx, searchRequest)
		resultChan <- AsyncResult[*response.Response[Doc]]{
			Result:  result,
			Error:   err,
			Context: searchRequest,
		}
	}()

	return resultChan
}

// BatchAsyncSearch 批量异步搜索
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
