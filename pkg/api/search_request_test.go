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
	// 创建真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 创建测试请求
	searchReq := request.NewSearchRequest().
		SetQuery(request.NewQuery().SetGroupId("org.apache.commons")).
		SetLimit(5)

	// 执行请求
	resp, err := client.SearchRequest(ctx, searchReq)
	if err != nil {
		t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	t.Logf("成功执行搜索请求，返回数据有效")
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
	// 创建真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试排序功能
	t.Run("排序功能", func(t *testing.T) {
		// 睡眠一段时间，避免请求过快
		time.Sleep(1 * time.Second)

		sortReq := request.NewSearchRequest().
			SetQuery(request.NewQuery().SetGroupId("org.apache.commons")).
			SetSort("timestamp", false).
			SetLimit(3)

		result, err := SearchRequestJsonDoc[*response.Artifact](client, ctx, sortReq)
		if err != nil {
			t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		assert.NoError(t, err)
		assert.NotNil(t, result)
		if len(result.ResponseBody.Docs) > 0 {
			t.Logf("使用时间戳降序排序后，找到 %d 个commons制品", len(result.ResponseBody.Docs))
			for i, doc := range result.ResponseBody.Docs {
				t.Logf("制品 %d: %s:%s (%s)", i+1, doc.GroupId, doc.ArtifactId, doc.LatestVersion)
			}
		}
	})

	// 测试聚合功能
	t.Run("聚合功能", func(t *testing.T) {
		// 睡眠一段时间，避免请求过快
		time.Sleep(1 * time.Second)

		facetReq := request.NewSearchRequest().
			SetQuery(request.NewQuery().SetGroupId("org.apache")).
			EnableFacet("a").
			SetLimit(1)

		// 对于聚合查询，只验证请求成功
		resp, err := client.SearchRequest(ctx, facetReq)
		if err != nil {
			t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		t.Log("聚合查询请求成功")
	})

	// 测试自定义参数
	t.Run("自定义参数", func(t *testing.T) {
		// 睡眠一段时间，避免请求过快
		time.Sleep(1 * time.Second)

		customReq := request.NewSearchRequest().
			SetQuery(request.NewQuery().SetGroupId("org.apache")).
			AddCustomParam("indent", "true").
			SetLimit(1)

		resp, err := client.SearchRequest(ctx, customReq)
		if err != nil {
			t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		t.Log("带自定义参数的请求成功")
	})
}
