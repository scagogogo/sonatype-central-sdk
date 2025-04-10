package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

func TestSearchRequestReal(t *testing.T) {
	// 创建一个真实的客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建搜索请求，查找log4j相关组件
	searchReq := request.NewSearchRequest()
	searchReq.Query.SetText("log4j")
	searchReq.SetLimit(5)

	// 执行搜索请求
	var result response.Response[*response.Artifact]
	err := client.SearchRequest(ctx, searchReq, &result)
	if err != nil {
		t.Fatalf("搜索请求失败: %v", err)
	}

	// 验证结果
	assert.Greater(t, result.ResponseHeader.QTime, 0)
	assert.NotNil(t, result.ResponseBody)
	assert.NotNil(t, result.ResponseBody.Docs)

	// 输出结果
	t.Logf("找到 %d 个结果", result.ResponseBody.NumFound)
	for i, doc := range result.ResponseBody.Docs {
		t.Logf("%d. %s:%s (%s)", i+1, doc.GroupId, doc.ArtifactId, doc.LatestVersion)
	}
}

func TestSearchRequestJsonDocReal(t *testing.T) {
	// 创建真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 创建测试请求
	searchReq := request.NewSearchRequest().
		SetQuery(request.NewQuery().SetGroupId("junit").SetArtifactId("junit")).
		SetLimit(3)

	// 执行请求并解析JSON
	result, err := SearchRequestJsonDoc[*response.Artifact](client, ctx, searchReq)
	if err != nil {
		t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.ResponseHeader)
	assert.NotNil(t, result.ResponseBody)

	// 验证文档内容
	if len(result.ResponseBody.Docs) > 0 {
		t.Logf("找到 %d 个junit制品", result.ResponseBody.NumFound)
		for i, doc := range result.ResponseBody.Docs[:minInt(3, len(result.ResponseBody.Docs))] {
			t.Logf("制品 %d: %s:%s (%s)", i+1, doc.GroupId, doc.ArtifactId, doc.LatestVersion)
			assert.Equal(t, "junit", doc.GroupId)
			assert.Equal(t, "junit", doc.ArtifactId)
			assert.NotEmpty(t, doc.LatestVersion)
		}
	} else {
		t.Log("未找到任何junit制品，这可能是API限制导致的")
	}
}

func TestSearchRequestWithAdvancedOptionsReal(t *testing.T) {
	// 创建一个真实的客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建一个请求，测试排序功能
	sortReq := request.NewSearchRequest()
	sortReq.Query.SetText("security library")
	sortReq.SetLimit(10)
	sortReq.SetSort("timestamp", false) // 按时间戳降序排序

	// 执行搜索请求
	var sortResult response.Response[*response.Artifact]
	err := client.SearchRequest(ctx, sortReq, &sortResult)
	if err != nil {
		t.Fatalf("排序搜索请求失败: %v", err)
	}

	// 验证结果
	assert.Greater(t, sortResult.ResponseBody.NumFound, int64(0))
	assert.LessOrEqual(t, len(sortResult.ResponseBody.Docs), 10)

	// 输出结果
	t.Log("=== 按时间戳排序的结果 ===")
	for i, doc := range sortResult.ResponseBody.Docs {
		t.Logf("%d. %s:%s (时间戳: %d)", i+1, doc.GroupId, doc.ArtifactId, doc.Timestamp)
		// 如果有超过1个结果，验证降序排序是否正确
		if i > 0 && len(sortResult.ResponseBody.Docs) > 1 {
			prev := sortResult.ResponseBody.Docs[i-1]
			assert.GreaterOrEqual(t, prev.Timestamp, doc.Timestamp, "结果应该按时间戳降序排序")
		}
	}

	// 测试使用自定义参数
	customReq := request.NewSearchRequest()
	customReq.Query.SetText("spring boot")
	customReq.AddCustomParam("fl", "id,g,a,latestVersion,p,timestamp")
	customReq.SetLimit(5)

	// 执行搜索请求
	var customResult response.Response[*response.Artifact]
	err = client.SearchRequest(ctx, customReq, &customResult)
	if err != nil {
		t.Fatalf("自定义参数搜索请求失败: %v", err)
	}

	// 验证结果
	assert.Greater(t, customResult.ResponseBody.NumFound, int64(0))
	t.Logf("使用自定义参数查询找到 %d 个结果", customResult.ResponseBody.NumFound)
}
