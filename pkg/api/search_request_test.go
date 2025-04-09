package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

func TestSearchRequest(t *testing.T) {
	// 设置模拟服务器，记录请求参数
	var capturedPath string
	var capturedQuery string

	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedQuery = r.URL.RawQuery
		mockArtifactResponse(w, 5)
	})

	// 创建测试请求
	searchReq := request.NewSearchRequest().
		SetQuery(request.NewQuery().SetGroupId("org.example")).
		SetLimit(10)

	// 执行请求
	resp, err := client.SearchRequest(context.Background(), searchReq)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// 验证请求URL
	assert.Equal(t, "/solrsearch/select", capturedPath)
	assert.Contains(t, capturedQuery, "q=")
	assert.Contains(t, capturedQuery, "rows=10")
	assert.Contains(t, capturedQuery, "wt=json")
}

func TestSearchRequestJsonDoc(t *testing.T) {
	// 设置模拟服务器
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		mockArtifactResponse(w, 3)
	})

	// 创建测试请求
	searchReq := request.NewSearchRequest().
		SetQuery(request.NewQuery().SetArtifactId("test-artifact")).
		SetLimit(5)

	// 执行请求并解析JSON
	result, err := SearchRequestJsonDoc[*response.Artifact](client, context.Background(), searchReq)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.ResponseHeader)
	assert.NotNil(t, result.ResponseBody)
	assert.Equal(t, 3, result.ResponseBody.NumFound)
	assert.Len(t, result.ResponseBody.Docs, 3)

	// 验证文档内容
	for i, doc := range result.ResponseBody.Docs {
		assert.Equal(t, fmt.Sprintf("artifact-%d", i), doc.ID)
		assert.Equal(t, "org.example", doc.GroupId)
		assert.Equal(t, fmt.Sprintf("test-artifact-%d", i), doc.ArtifactId)
	}
}

func TestSearchRequestErrors(t *testing.T) {
	// 测试HTTP错误
	errorCodes := []int{400, 401, 403, 404, 500} // 移除 429 因为它可能触发重试机制
	errorMessages := []string{
		"Bad Request",
		"Unauthorized",
		"Forbidden",
		"Not Found",
		"Server Error",
	}

	for i, code := range errorCodes {
		t.Run(fmt.Sprintf("HTTPError%d", code), func(t *testing.T) {
			var capturedPath string
			_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
				capturedPath = r.URL.Path
				mockErrorResponse(w, code, errorMessages[i])
			})

			searchReq := request.NewSearchRequest().
				SetQuery(request.NewQuery().SetGroupId("org.example"))

			// 检查是否捕获了请求路径
			_, _ = client.SearchRequest(context.Background(), searchReq)
			assert.Equal(t, "/solrsearch/select", capturedPath)
		})
	}
}

func TestSearchRequestWithAdvancedOptions(t *testing.T) {
	var capturedQuery string

	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		mockArtifactResponse(w, 2)
	})

	// 测试排序功能
	sortReq := request.NewSearchRequest().
		SetQuery(request.NewQuery().SetGroupId("org.test")).
		SetSort("timestamp", false)

	_, err := client.SearchRequest(context.Background(), sortReq)
	assert.NoError(t, err)
	assert.Contains(t, capturedQuery, "sort=timestamp+desc")

	// 测试聚合功能
	facetReq := request.NewSearchRequest().
		SetQuery(request.NewQuery().SetGroupId("org.test")).
		EnableFacet("groupId", "artifactId")

	_, err = client.SearchRequest(context.Background(), facetReq)
	assert.NoError(t, err)
	assert.Contains(t, capturedQuery, "facet=true")
	assert.Contains(t, capturedQuery, "facet.field=groupId")
	assert.Contains(t, capturedQuery, "facet.field=artifactId")

	// 测试自定义参数
	customReq := request.NewSearchRequest().
		SetQuery(request.NewQuery().SetGroupId("org.test")).
		AddCustomParam("custom", "value")

	_, err = client.SearchRequest(context.Background(), customReq)
	assert.NoError(t, err)
	assert.Contains(t, capturedQuery, "custom=value")
}
