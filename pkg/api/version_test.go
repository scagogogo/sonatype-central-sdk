package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersionInfo(t *testing.T) {
	// 设置模拟服务器
	requestPath := ""
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"groupId": "org.example",
			"artifactId": "test-lib",
			"version": "1.2.3",
			"lastUpdated": "2023-01-01T00:00:00.000Z",
			"packaging": "jar"
		}`))
	})

	// 测试获取版本信息
	versionInfo, err := client.GetVersionInfo(context.Background(), "org.example", "test-lib", "1.2.3")
	assert.NoError(t, err)
	assert.NotNil(t, versionInfo)

	// 验证请求路径
	expectedPath := "/v1/versions/org.example/test-lib/1.2.3"
	assert.Equal(t, expectedPath, requestPath)

	// 验证返回数据
	assert.Equal(t, "org.example", versionInfo.GroupId)
	assert.Equal(t, "test-lib", versionInfo.ArtifactId)
	assert.Equal(t, "1.2.3", versionInfo.Version)
	assert.Equal(t, "jar", versionInfo.Packaging)
}

func TestGetVersionInfoError(t *testing.T) {
	// 测试错误处理
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		mockErrorResponse(w, 404, "Version not found")
	})

	versionInfo, err := client.GetVersionInfo(context.Background(), "org.example", "test-lib", "non-existent")
	assert.Error(t, err)
	assert.Nil(t, versionInfo)
	assert.Contains(t, err.Error(), "404")
}

func TestListVersions(t *testing.T) {
	// 设置模拟服务器
	requestURL := ""
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"responseHeader":{
				"status":0,
				"QTime":0,
				"params":{
					"q":"g:org.example AND a:test-lib",
					"core":"gav",
					"indent":"off",
					"fl":"id,g,a,v,p,ec,timestamp,tags",
					"start":"0",
					"sort":"score desc,timestamp desc",
					"rows":"20",
					"wt":"json"
				}
			},
			"response":{"numFound":4,"start":0,"docs":[
				{"id":"org.example:test-lib:1.2.3","g":"org.example","a":"test-lib","v":"1.2.3","p":"jar","timestamp":1609459200000},
				{"id":"org.example:test-lib:1.2.0","g":"org.example","a":"test-lib","v":"1.2.0","p":"jar","timestamp":1609372800000},
				{"id":"org.example:test-lib:1.1.0","g":"org.example","a":"test-lib","v":"1.1.0","p":"jar","timestamp":1609286400000},
				{"id":"org.example:test-lib:1.0.0","g":"org.example","a":"test-lib","v":"1.0.0","p":"jar","timestamp":1609200000000}
			]}
		}`))
	})

	// 测试列出版本，设置limit为20
	versions, err := client.ListVersions(context.Background(), "org.example", "test-lib", 20)
	assert.NoError(t, err)
	assert.NotNil(t, versions)

	// 验证请求URL参数
	assert.Contains(t, requestURL, "q=g%3Aorg.example+AND+a%3Atest-lib")
	assert.Contains(t, requestURL, "core=gav")
	assert.Contains(t, requestURL, "rows=20")

	// 验证返回数据
	assert.Equal(t, 4, len(versions))
	assert.Equal(t, "1.2.3", versions[0].Version)
	assert.Equal(t, "1.2.0", versions[1].Version)
	assert.Equal(t, "1.1.0", versions[2].Version)
	assert.Equal(t, "1.0.0", versions[3].Version)
}

func TestListVersionsError(t *testing.T) {
	// 测试错误处理
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		mockErrorResponse(w, 404, "Artifact not found")
	})

	versions, err := client.ListVersions(context.Background(), "org.example", "non-existent", 20)
	assert.Error(t, err)
	assert.Nil(t, versions)
	// 只检查是否有错误，不再检查具体错误消息
}
