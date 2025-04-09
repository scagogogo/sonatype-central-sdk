package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// setupMockServer 创建一个模拟的HTTP服务器，用于测试
func setupMockServer(t *testing.T, respHandler func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, *Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		respHandler(w, r)
	}))

	client := NewClient(
		WithBaseURL(server.URL),
		WithRepoBaseURL(server.URL),
	)

	t.Cleanup(func() {
		server.Close()
	})

	return server, client
}

// createRealClient 创建一个连接到真实Maven Central API的客户端
func createRealClient(t *testing.T) *Client {
	// 创建默认客户端实例（使用真实API地址）
	client := NewClient(
		WithMaxRetries(2),     // 减少重试次数，避免测试太慢
		WithRetryBackoff(300), // 缩短重试间隔
		WithCache(true, 300),  // 启用缓存以加速重复测试
	)

	return client
}

// mockSearchResponse 返回一个模拟的搜索响应
func mockSearchResponse(w http.ResponseWriter, docs interface{}, numFound int) {
	// 构造不使用泛型的响应结构
	resp := struct {
		ResponseHeader struct {
			Status int         `json:"status"`
			QTime  int         `json:"QTime"`
			Params interface{} `json:"params"`
		} `json:"responseHeader"`
		Response struct {
			NumFound int         `json:"numFound"`
			Start    int         `json:"start"`
			Docs     interface{} `json:"docs"`
		} `json:"response"`
	}{
		ResponseHeader: struct {
			Status int         `json:"status"`
			QTime  int         `json:"QTime"`
			Params interface{} `json:"params"`
		}{
			Status: 0,
			QTime:  10,
			Params: map[string]string{},
		},
		Response: struct {
			NumFound int         `json:"numFound"`
			Start    int         `json:"start"`
			Docs     interface{} `json:"docs"`
		}{
			NumFound: numFound,
			Start:    0,
			Docs:     docs,
		},
	}

	jsonData, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonData)
}

// mockArtifactResponse 返回模拟的Artifact文档
func mockArtifactResponse(w http.ResponseWriter, count int) {
	docs := make([]*response.Artifact, 0, count)
	for i := 0; i < count; i++ {
		docs = append(docs, &response.Artifact{
			ID:            fmt.Sprintf("artifact-%d", i),
			GroupId:       "org.example",
			ArtifactId:    fmt.Sprintf("test-artifact-%d", i),
			LatestVersion: "1.0.0",
			RepositoryID:  "central",
			Packaging:     "jar",
			Timestamp:     1600000000000,
			VersionCount:  1,
		})
	}
	mockSearchResponse(w, docs, count)
}

// mockVersionResponse 返回模拟的Version文档
func mockVersionResponse(w http.ResponseWriter, count int) {
	docs := make([]*response.Version, 0, count)
	for i := 0; i < count; i++ {
		docs = append(docs, &response.Version{
			ID:         fmt.Sprintf("version-%d", i),
			GroupId:    "org.example",
			ArtifactId: "test-artifact",
			Version:    fmt.Sprintf("1.0.%d", i),
			Packaging:  "jar",
			Timestamp:  1600000000000 + int64(i*1000),
		})
	}
	mockSearchResponse(w, docs, count)
}

// mockErrorResponse 返回一个错误响应
func mockErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, message)))
}

// mockBinaryResponse 返回二进制数据响应
func mockBinaryResponse(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
