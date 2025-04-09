package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchByGroupId(t *testing.T) {
	// 设置模拟服务器
	requestPath := ""
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path + "?" + r.URL.RawQuery
		mockArtifactResponse(w, 5)
	})

	// 测试搜索
	artifacts, err := client.SearchByGroupId(context.Background(), "org.example", 5)
	assert.NoError(t, err)
	assert.NotNil(t, artifacts)
	assert.Len(t, artifacts, 5)

	// 检查请求格式
	assert.Contains(t, requestPath, "q=g%3Aorg.example")
	assert.Contains(t, requestPath, "rows=5")

	// 检查返回结果
	for _, artifact := range artifacts {
		assert.Equal(t, "org.example", artifact.GroupId)
		assert.Contains(t, artifact.ArtifactId, "test-artifact-")
		assert.Equal(t, "1.0.0", artifact.LatestVersion)
	}
}

func TestSearchByGroupIdWithNoLimit(t *testing.T) {
	// 设置模拟服务器，模拟分页
	requestCount := 0
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		// 第一次请求返回最大数量，第二次请求返回空
		if requestCount == 1 {
			mockArtifactResponse(w, 200)
		} else {
			mockArtifactResponse(w, 0)
		}
	})

	// 测试无限制搜索 (会使用迭代器)
	artifacts, err := client.SearchByGroupId(context.Background(), "popular-group", 0)
	assert.NoError(t, err)
	assert.NotNil(t, artifacts)
	assert.Equal(t, 200, len(artifacts)) // 应该获取所有的结果
	assert.Equal(t, 1, requestCount)     // 由于空检查，当第一次请求已经获取到所有结果时，不会发送第二次请求
}

func TestSearchByGroupIdError(t *testing.T) {
	// 测试错误处理
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		mockErrorResponse(w, 500, "Internal Server Error")
	})

	artifacts, err := client.SearchByGroupId(context.Background(), "org.error", 10)
	assert.Error(t, err)
	assert.Nil(t, artifacts)
}

func TestIteratorByGroupId(t *testing.T) {
	// 设置模拟服务器，返回两次结果
	pageCount := 0
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		pageCount++
		if pageCount == 1 {
			// 第一页
			mockArtifactResponse(w, 5)
		} else {
			// 第二页，没有更多结果
			mockArtifactResponse(w, 0)
		}
	})

	// 创建迭代器
	iterator := client.IteratorByGroupId(context.Background(), "org.example")
	assert.NotNil(t, iterator)

	// 迭代所有元素
	count := 0
	for iterator.Next() {
		artifact := iterator.Value()
		assert.NotNil(t, artifact)
		assert.Equal(t, "org.example", artifact.GroupId)
		count++
	}

	assert.Equal(t, 5, count)
	assert.Equal(t, 1, pageCount) // 由于空检查，当第一次请求已经获取到所有结果时，不会发送第二次请求
}
