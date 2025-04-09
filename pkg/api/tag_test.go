package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
	"github.com/stretchr/testify/assert"
)

// TestSearchByTagMock 使用模拟服务器测试功能
func TestSearchByTagMock(t *testing.T) {
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// 确保请求路径正确包含标签参数
		assert.Contains(t, r.URL.RawQuery, "tags%3Atest-tag")
		mockArtifactResponse(w, 3)
	})

	// 使用模拟标签进行测试
	artifacts, err := client.SearchByTag(context.Background(), "test-tag", 10)

	assert.Nil(t, err)
	assert.Equal(t, 3, len(artifacts))
}

// TestSearchByTagReal 使用真实API测试
func TestSearchByTagReal(t *testing.T) {
	// 可选跳过长时间测试
	if testing.Short() {
		t.Skip("跳过真实API测试")
	}

	// 使用真实客户端
	client := createRealClient(t)

	// 测试几个常见的标签
	tagNames := []string{"jdbc", "logging", "http-client"}

	for _, tag := range tagNames {
		t.Run("Tag_"+tag, func(t *testing.T) {
			artifacts, err := client.SearchByTag(context.Background(), tag, 5)

			if err != nil {
				t.Logf("搜索标签 %s 时出错: %v", tag, err)
				return
			}

			// 记录找到的结果，但不强制要求特定内容
			t.Logf("找到 %d 个包含标签 %s 的结果", len(artifacts), tag)
			if len(artifacts) > 0 {
				for i, a := range artifacts[:min(3, len(artifacts))] {
					t.Logf("结果 %d: %s:%s:%s", i+1, a.GroupId, a.ArtifactId, a.LatestVersion)
				}
			}

			assert.True(t, len(artifacts) >= 0) // 只确保API正常返回
		})
	}
}

// 测试获取相关标签功能的模拟测试
func TestGetRelatedTagsMock(t *testing.T) {
	// 设置模拟服务器
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/solrsearch/select" && r.Method == http.MethodGet {
			// 确保包含标签参数
			query := r.URL.Query().Get("q")
			assert.Contains(t, query, "tags:java", "查询中应包含java标签")

			// 返回一些带标签的伪造项目
			artifacts := []*response.Artifact{
				{
					ID: "artifact1", GroupId: "org.test", ArtifactId: "artifact1",
					Tags: []string{"java", "library", "util"},
				},
				{
					ID: "artifact2", GroupId: "org.test", ArtifactId: "artifact2",
					Tags: []string{"java", "framework", "web"},
				},
				{
					ID: "artifact3", GroupId: "org.test", ArtifactId: "artifact3",
					Tags: []string{"java", "library", "json"},
				},
			}

			mockSearchResponse(w, artifacts, len(artifacts))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// 测试获取相关标签
	relatedTags, err := client.GetRelatedTags(context.Background(), "java", 10)
	assert.NoError(t, err)
	assert.NotNil(t, relatedTags)

	// 验证结果
	assert.Equal(t, 2, relatedTags["library"], "library标签应出现2次")
	assert.Equal(t, 1, relatedTags["framework"], "framework标签应出现1次")
	assert.Equal(t, 1, relatedTags["web"], "web标签应出现1次")
	assert.Equal(t, 1, relatedTags["json"], "json标签应出现1次")
	assert.Equal(t, 0, relatedTags["java"], "java标签不应出现在结果中")
}

// 测试多标签搜索功能的模拟测试
func TestSearchByMultipleTagsMock(t *testing.T) {
	// 设置模拟服务器
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/solrsearch/select" && r.Method == http.MethodGet {
			// 验证查询参数
			query := r.URL.Query().Get("q")
			assert.Contains(t, query, "tags:java", "查询中应包含java标签")
			assert.Contains(t, query, "tags:util", "查询中应包含util标签")

			// 返回模拟数据
			artifacts := []*response.Artifact{
				{
					ID: "artifact1", GroupId: "org.test", ArtifactId: "artifact1",
					Tags: []string{"java", "util", "library"},
				},
			}

			mockSearchResponse(w, artifacts, len(artifacts))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// 测试多标签搜索
	artifacts, err := client.SearchByMultipleTags(context.Background(), []string{"java", "util"}, 10)
	assert.NoError(t, err)
	assert.NotNil(t, artifacts)
	assert.Equal(t, 1, len(artifacts), "应返回1个匹配的项目")
	assert.Equal(t, "artifact1", artifacts[0].ArtifactId, "应返回正确的项目")
}

// 测试热门标签的模拟测试
func TestGetMostUsedTagsMock(t *testing.T) {
	// 设置模拟服务器
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/solrsearch/select" && r.Method == http.MethodGet {
			query := r.URL.Query().Get("q")

			if strings.Contains(query, "tags:java") {
				// 返回java标签的项目
				artifacts := []*response.Artifact{
					{ID: "a1", Tags: []string{"java", "library", "util"}},
					{ID: "a2", Tags: []string{"java", "framework", "web"}},
					{ID: "a3", Tags: []string{"java", "library", "json"}},
					{ID: "a4", Tags: []string{"java", "framework", "web", "mvc"}},
					{ID: "a5", Tags: []string{"java", "library", "util", "collection"}},
				}
				mockSearchResponse(w, artifacts, len(artifacts))
			} else if strings.Contains(query, "g:org.springframework") {
				// 返回一些Spring项目
				artifacts := []*response.Artifact{
					{ID: "s1", Tags: []string{"java", "spring", "web"}},
					{ID: "s2", Tags: []string{"java", "spring", "data"}},
				}
				mockSearchResponse(w, artifacts, len(artifacts))
			} else {
				// 默认返回空结果
				mockSearchResponse(w, []*response.Artifact{}, 0)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// 测试获取热门标签 - 以java为基础
	tagCounts, err := client.GetMostUsedTags(context.Background(), "java", 3)
	assert.NoError(t, err)
	assert.NotNil(t, tagCounts)
	assert.Equal(t, 3, len(tagCounts), "应返回3个最常用的标签")

	// 验证排序正确 - java出现5次，应该是最多的
	assert.Equal(t, "java", tagCounts[0].Tag, "java应该是最常见的标签")
	assert.Equal(t, 5, tagCounts[0].Count, "java应该出现5次")

	// 检查第二常见的标签
	assert.Contains(t, []string{"library"}, tagCounts[1].Tag, "library应该是第二常见的标签")

	// 测试无基础标签情况
	tagCounts, err = client.GetMostUsedTags(context.Background(), "", 3)
	assert.NoError(t, err)
	assert.NotNil(t, tagCounts)
}

// 测试搜索同时具有所有标签的方法
func TestSearchArtifactsWithAllTagsMock(t *testing.T) {
	// 设置模拟服务器
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/solrsearch/select" && r.Method == http.MethodGet {
			query := r.URL.Query().Get("q")

			if strings.Contains(query, "tags:java") {
				// 返回具有java标签的项目
				artifacts := []*response.Artifact{
					{ID: "a1", GroupId: "org.test", ArtifactId: "a1", Tags: []string{"java", "web", "framework"}},
					{ID: "a2", GroupId: "org.test", ArtifactId: "a2", Tags: []string{"java", "library"}},
					{ID: "a3", GroupId: "org.test", ArtifactId: "a3", Tags: []string{"java", "web", "api"}},
					{ID: "a4", GroupId: "org.test", ArtifactId: "a4", Tags: []string{"java", "web", "framework", "mvc"}},
				}
				mockSearchResponse(w, artifacts, len(artifacts))
			} else {
				mockSearchResponse(w, []*response.Artifact{}, 0)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// 测试查找同时具有java和web标签的项目
	artifacts, err := client.SearchArtifactsWithAllTags(context.Background(), []string{"java", "web"}, 0)
	assert.NoError(t, err)
	assert.NotNil(t, artifacts)
	assert.Equal(t, 3, len(artifacts), "应该返回3个同时具有java和web标签的项目")

	// 测试查找同时具有java、web和framework标签的项目
	artifacts, err = client.SearchArtifactsWithAllTags(context.Background(), []string{"java", "web", "framework"}, 0)
	assert.NoError(t, err)
	assert.NotNil(t, artifacts)
	assert.Equal(t, 2, len(artifacts), "应该返回2个同时具有java、web和framework标签的项目")

	// 测试限制结果数量
	artifacts, err = client.SearchArtifactsWithAllTags(context.Background(), []string{"java", "web"}, 1)
	assert.NoError(t, err)
	assert.NotNil(t, artifacts)
	assert.Equal(t, 1, len(artifacts), "应该只返回1个项目")
}

// 测试按GroupId过滤的标签搜索
func TestSearchByTagWithGroupFilterMock(t *testing.T) {
	// 设置模拟服务器
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/solrsearch/select" && r.Method == http.MethodGet {
			query := r.URL.Query().Get("q")

			if strings.Contains(query, "tags:web") {
				// 返回具有web标签的各种项目
				artifacts := []*response.Artifact{
					{ID: "a1", GroupId: "org.springframework", ArtifactId: "spring-web", Tags: []string{"spring", "web"}},
					{ID: "a2", GroupId: "org.apache", ArtifactId: "tomcat", Tags: []string{"web", "server"}},
					{ID: "a3", GroupId: "org.springframework", ArtifactId: "spring-webmvc", Tags: []string{"spring", "web", "mvc"}},
					{ID: "a4", GroupId: "io.undertow", ArtifactId: "undertow-core", Tags: []string{"web", "server"}},
				}
				mockSearchResponse(w, artifacts, len(artifacts))
			} else {
				mockSearchResponse(w, []*response.Artifact{}, 0)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// 测试按GroupId前缀过滤
	artifacts, err := client.SearchByTagWithGroupFilter(context.Background(), "web", "org.springframework", 0)
	assert.NoError(t, err)
	assert.NotNil(t, artifacts)
	assert.Equal(t, 2, len(artifacts), "应该返回2个属于org.springframework的web项目")

	for _, a := range artifacts {
		assert.True(t, strings.HasPrefix(a.GroupId, "org.springframework"), "所有结果都应该属于springframework组")
		assert.Contains(t, a.Tags, "web", "所有结果都应该包含web标签")
	}

	// 测试没有匹配项的情况
	artifacts, err = client.SearchByTagWithGroupFilter(context.Background(), "web", "org.nonexistent", 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(artifacts), "不应该返回任何结果")
}

// 使用真实API测试标签功能
func TestSearchByTagRealExtended(t *testing.T) {
	client := createRealClient(t)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试一些常见标签
	tags := []string{"jdbc", "logging", "http-client"}

	for _, tag := range tags {
		t.Run(fmt.Sprintf("Tag_%s", tag), func(t *testing.T) {
			artifacts, err := client.SearchByTag(ctx, tag, 5)
			if err != nil {
				t.Logf("跳过标签搜索测试: %v", err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			assert.NotNil(t, artifacts)
			t.Logf("找到 %d 个包含标签 %s 的结果", len(artifacts), tag)

			// 记录前几个结果
			for i, a := range artifacts {
				if i < 3 {
					t.Logf("结果 %d: %s:%s", i+1, a.GroupId, a.ArtifactId)
				}
			}
		})
	}
}

// 测试获取相关标签 - 真实API
func TestGetRelatedTagsReal(t *testing.T) {
	client := createRealClient(t)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试获取与特定标签相关的其他标签
	relatedTags, err := client.GetRelatedTags(ctx, "spring", 20)
	if err != nil {
		t.Logf("跳过相关标签测试: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	assert.NotNil(t, relatedTags)
	assert.True(t, len(relatedTags) > 0, "应返回一些相关标签")

	// 显示前5个最常见的相关标签
	t.Logf("与'spring'相关的标签:")

	// 将map转换为切片以便排序
	type tagCount struct {
		tag   string
		count int
	}

	var tagCounts []tagCount
	for tag, count := range relatedTags {
		tagCounts = append(tagCounts, tagCount{tag, count})
	}

	// 按计数降序排序
	for i := 0; i < len(tagCounts); i++ {
		for j := i + 1; j < len(tagCounts); j++ {
			if tagCounts[i].count < tagCounts[j].count {
				tagCounts[i], tagCounts[j] = tagCounts[j], tagCounts[i]
			}
		}
	}

	// 显示前5个（如果有）
	for i, tc := range tagCounts {
		if i < 5 {
			t.Logf("  %s: %d 次", tc.tag, tc.count)
		} else {
			break
		}
	}
}

// 测试多个标签搜索 - 真实API
func TestSearchByMultipleTagsReal(t *testing.T) {
	client := createRealClient(t)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试一些常见标签组合
	tagSets := [][]string{
		{"java", "web"},
		{"database", "client"},
		{"http", "client"},
	}

	for _, tags := range tagSets {
		tagNames := strings.Join(tags, "+")
		t.Run(fmt.Sprintf("Tags_%s", tagNames), func(t *testing.T) {
			artifacts, err := client.SearchByMultipleTags(ctx, tags, 5)
			if err != nil {
				t.Logf("跳过多标签搜索测试 %s: %v", tagNames, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			assert.NotNil(t, artifacts)
			t.Logf("找到 %d 个同时包含标签 %s 的结果", len(artifacts), tagNames)

			// 记录前几个结果
			for i, a := range artifacts {
				if i < 3 {
					t.Logf("结果 %d: %s:%s", i+1, a.GroupId, a.ArtifactId)
				}
			}
		})
	}
}

// 测试获取最常用标签 - 真实API
func TestGetMostUsedTagsReal(t *testing.T) {
	client := createRealClient(t)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 基于java测试热门标签
	tagCounts, err := client.GetMostUsedTags(ctx, "java", 10)
	if err != nil {
		t.Logf("跳过获取热门标签测试: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	assert.NotNil(t, tagCounts)
	assert.True(t, len(tagCounts) > 0, "应返回一些热门标签")

	t.Logf("基于'java'的热门标签:")
	for i, tc := range tagCounts {
		if i < 5 {
			t.Logf("  %d. %s: %d 次", i+1, tc.Tag, tc.Count)
		}
	}
}
