package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// TestListVersionsReal 使用真实 API 测试获取版本列表功能
func TestListVersionsReal(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试真实API - 使用常见的依赖项进行测试
	versions, err := client.ListVersions(ctx, "com.google.inject", "guice", 20)
	if err != nil {
		t.Logf("跳过测试：无法获取guice版本列表: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	assert.NotNil(t, versions)
	assert.True(t, len(versions) > 0, "应该返回多个版本")

	// 验证至少返回了几个常见版本
	var hasVersion4 bool
	for _, v := range versions {
		if v.Version == "4.0" {
			hasVersion4 = true
			break
		}
	}
	assert.True(t, hasVersion4, "应该包含4.0版本")

	// 记录找到的版本供参考
	t.Logf("找到 %d 个 guice 版本", len(versions))
	for i, v := range versions[:min(5, len(versions))] {
		t.Logf("版本 %d: %s", i+1, v.Version)
	}
}

// 测试另一个常见依赖的版本列表
func TestListVersionsRealCommonLib(t *testing.T) {
	client := createRealClient(t)

	// 测试 Apache Commons 等常用库
	libraries := []struct {
		groupId    string
		artifactId string
	}{
		{"org.apache.commons", "commons-lang3"},
		{"org.slf4j", "slf4j-api"},
		{"com.google.code.gson", "gson"},
	}

	for _, lib := range libraries {
		t.Run(lib.groupId+":"+lib.artifactId, func(t *testing.T) {
			// 设置超时上下文
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			versions, err := client.ListVersions(ctx, lib.groupId, lib.artifactId, 5)
			if err != nil {
				t.Logf("跳过测试 %s:%s: %v", lib.groupId, lib.artifactId, err)
				t.Skip("无法连接到Maven Central API")
				return
			}
			assert.NotEmpty(t, versions, "应该找到版本")

			t.Logf("找到 %s:%s 的 %d 个版本", lib.groupId, lib.artifactId, len(versions))
			for _, v := range versions {
				t.Logf("  - %s", v.Version)
			}
		})
	}
}

// 测试获取最新版本功能
func TestGetLatestVersionReal(t *testing.T) {
	client := createRealClient(t)

	// 测试一些广泛使用的库
	testCases := []struct {
		name       string
		groupId    string
		artifactId string
	}{
		{"Guice", "com.google.inject", "guice"},
		{"JUnit", "junit", "junit"},
		{"Spring Core", "org.springframework", "spring-core"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 设置超时
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			latestVersion, err := client.GetLatestVersion(ctx, tc.groupId, tc.artifactId)
			if err != nil {
				t.Logf("跳过获取最新版本测试 %s:%s: %v", tc.groupId, tc.artifactId, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			assert.NotNil(t, latestVersion)
			assert.NotEmpty(t, latestVersion.Version)

			t.Logf("%s:%s 最新版本: %s", tc.groupId, tc.artifactId, latestVersion.Version)
		})
	}
}

// 测试版本过滤功能
func TestFilterVersionsReal(t *testing.T) {
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试过滤器 - 只获取版本号以"4."开头的版本
	versions, err := client.FilterVersions(ctx, "com.google.inject", "guice", func(v *response.Version) bool {
		return strings.HasPrefix(v.Version, "4.")
	})

	if err != nil {
		t.Logf("跳过版本过滤测试: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	assert.NotNil(t, versions)

	if len(versions) > 0 {
		t.Logf("找到 %d 个以'4.'开头的Guice版本", len(versions))
		for _, v := range versions {
			t.Logf("  - %s", v.Version)
			assert.True(t, strings.HasPrefix(v.Version, "4."), "版本应该以'4.'开头")
		}
	} else {
		t.Log("没有找到以'4.'开头的版本，这可能是合理的")
	}
}

// 测试获取带元数据的版本列表
func TestGetVersionsWithMetadataReal(t *testing.T) {
	client := createRealClient(t)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 使用查询限制，以避免请求太多版本，这可能会导致测试很慢
	allVersions, err := client.ListVersions(ctx, "junit", "junit", 3)
	if err != nil || len(allVersions) == 0 {
		t.Logf("跳过测试：无法获取junit版本列表: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	versionsWithMeta, err := client.GetVersionsWithMetadata(ctx, "junit", "junit")
	if err != nil {
		t.Logf("跳过测试：获取版本元数据失败: %v", err)
		t.Skip("无法获取版本元数据")
		return
	}

	if len(versionsWithMeta) > 0 {
		t.Logf("找到 %d 个junit版本带元数据", len(versionsWithMeta))
		for i, v := range versionsWithMeta[:min(3, len(versionsWithMeta))] {
			t.Logf("版本 %d: %s", i+1, v.Version.Version)
			assert.NotNil(t, v.Version, "版本不应为空")
			assert.NotNil(t, v.VersionInfo, "版本信息不应为空")
			assert.Equal(t, v.Version.Version, v.VersionInfo.Version, "版本号应匹配")
			assert.NotEmpty(t, v.VersionInfo.LastUpdated, "最后更新时间不应为空")
		}
	} else {
		t.Log("没有找到带元数据的版本")
	}
}

// 测试版本比较功能
func TestCompareVersionsReal(t *testing.T) {
	client := createRealClient(t)

	// 设置测试超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 比较两个已知存在的JUnit版本
	comparison, err := client.CompareVersions(ctx, "junit", "junit", "4.12", "4.13")
	if err != nil {
		t.Logf("跳过比较版本测试: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}
	assert.NotNil(t, comparison)

	t.Logf("比较版本: %s vs %s", comparison.Version1, comparison.Version2)
	t.Logf("版本1时间戳: %s", comparison.V1Timestamp)
	t.Logf("版本2时间戳: %s", comparison.V2Timestamp)

	assert.Equal(t, "4.12", comparison.Version1)
	assert.Equal(t, "4.13", comparison.Version2)
	assert.NotEmpty(t, comparison.V1Timestamp)
	assert.NotEmpty(t, comparison.V2Timestamp)
}

// 测试版本存在性检查
func TestHasVersionReal(t *testing.T) {
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试一个肯定存在的版本
	exists, err := client.HasVersion(ctx, "junit", "junit", "4.12")
	if err != nil {
		t.Logf("跳过版本存在性检查测试: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}
	assert.True(t, exists, "junit:junit:4.12应该存在")

	// 测试一个肯定不存在的版本
	exists, err = client.HasVersion(ctx, "junit", "junit", "999.999.999")
	assert.NoError(t, err, "检查不存在的版本应该不返回错误")
	assert.False(t, exists, "junit:junit:999.999.999不应该存在")

	// 随机生成的不存在的GroupID
	randomID := fmt.Sprintf("org.nonexistent.test.%d", time.Now().UnixNano())
	exists, err = client.HasVersion(ctx, randomID, "nonexistent", "1.0")
	assert.NoError(t, err, "检查不存在的版本应该不返回错误")
	assert.False(t, exists, fmt.Sprintf("%s:nonexistent:1.0不应该存在", randomID))
}

// 测试使用模拟服务器的GetLatestVersion功能
func TestGetLatestVersionMock(t *testing.T) {
	// 设置模拟服务器
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/solrsearch/select" && r.Method == http.MethodGet {
			// 检查查询参数
			query := r.URL.Query()

			// 确保设置了正确的参数
			assert.Equal(t, "gav", query.Get("core"), "core参数应为gav")
			assert.Equal(t, "1", query.Get("rows"), "rows参数应为1")
			assert.Equal(t, "g:test.group AND a:test.artifact", query.Get("q"), "应查询正确的groupId和artifactId")

			// 返回模拟数据
			mockVersionResponse(w, 1)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// 测试GetLatestVersion
	latestVersion, err := client.GetLatestVersion(context.Background(), "test.group", "test.artifact")
	assert.NoError(t, err)
	assert.NotNil(t, latestVersion)
	assert.Contains(t, latestVersion.Version, "1.0.", "应返回由mockVersionResponse生成的版本")
}

// 测试使用模拟服务器的FilterVersions功能
func TestFilterVersionsMock(t *testing.T) {
	// 设置模拟服务器
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/solrsearch/select" && r.Method == http.MethodGet {
			// 检查查询参数
			query := r.URL.Query()

			// 确保设置了正确的参数
			assert.Equal(t, "gav", query.Get("core"), "core参数应为gav")
			assert.Equal(t, "g:org.test AND a:test-lib", query.Get("q"), "应查询正确的groupId和artifactId")

			// 返回多个测试版本
			versions := []*response.Version{
				{Version: "1.0.0", GroupId: "org.test", ArtifactId: "test-lib"},
				{Version: "1.1.0", GroupId: "org.test", ArtifactId: "test-lib"},
				{Version: "2.0.0", GroupId: "org.test", ArtifactId: "test-lib"},
				{Version: "2.1.0", GroupId: "org.test", ArtifactId: "test-lib"},
			}

			mockSearchResponse(w, versions, len(versions))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// 测试过滤器 - 只选择2.x版本
	filteredVersions, err := client.FilterVersions(context.Background(), "org.test", "test-lib", func(v *response.Version) bool {
		return strings.HasPrefix(v.Version, "2.")
	})

	assert.NoError(t, err)
	assert.NotNil(t, filteredVersions)
	assert.Equal(t, 2, len(filteredVersions), "应该找到2个2.x版本")

	for _, v := range filteredVersions {
		assert.True(t, strings.HasPrefix(v.Version, "2."), "所有版本都应以2.开头")
	}
}

// 测试使用模拟服务器的HasVersion功能
func TestHasVersionMock(t *testing.T) {
	t.Run("版本存在", func(t *testing.T) {
		// 设置模拟服务器 - 正常响应
		_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/versions/test.group/test.artifact/1.0.0" && r.Method == http.MethodGet {
				// 返回一个有效的版本信息
				versionInfo := response.VersionInfo{
					GroupId:     "test.group",
					ArtifactId:  "test.artifact",
					Version:     "1.0.0",
					LastUpdated: "2023-01-01T10:10:10.000Z",
					Packaging:   "jar",
				}
				json.NewEncoder(w).Encode(versionInfo)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		})

		// 测试存在的版本
		exists, err := client.HasVersion(context.Background(), "test.group", "test.artifact", "1.0.0")
		assert.NoError(t, err)
		assert.True(t, exists, "版本应该被报告为存在")
	})

	t.Run("版本不存在", func(t *testing.T) {
		// 设置模拟服务器 - 404响应
		_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/versions/test.group/test.artifact/999.0.0" && r.Method == http.MethodGet {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error":"Version not found"}`))
			}
		})

		// 测试不存在的版本
		exists, err := client.HasVersion(context.Background(), "test.group", "test.artifact", "999.0.0")
		assert.NoError(t, err)
		assert.False(t, exists, "版本应该被报告为不存在")
	})
}

// 测试多种条件的版本过滤
func TestFilterVersionsWithMultipleConditions(t *testing.T) {
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 使用多个条件过滤版本 - 找出2.0以上但不包含beta/alpha的稳定版本
	versions, err := client.FilterVersions(ctx, "com.google.code.gson", "gson", func(v *response.Version) bool {
		// 检查版本是否大于2.0
		if !strings.HasPrefix(v.Version, "2.") && !strings.HasPrefix(v.Version, "3.") {
			return false
		}

		// 排除beta、alpha、RC等非稳定版本
		if strings.Contains(strings.ToLower(v.Version), "beta") ||
			strings.Contains(strings.ToLower(v.Version), "alpha") ||
			strings.Contains(strings.ToLower(v.Version), "rc") ||
			strings.Contains(strings.ToLower(v.Version), "snapshot") {
			return false
		}

		return true
	})

	if err != nil {
		t.Logf("跳过多条件版本过滤测试: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	assert.NotNil(t, versions)
	assert.True(t, len(versions) > 0, "应该找到符合条件的版本")

	// 验证所有返回的版本都符合我们的过滤条件
	for _, v := range versions {
		t.Logf("过滤后的版本: %s", v.Version)

		// 应该以2.或3.开头
		assert.True(t, strings.HasPrefix(v.Version, "2.") || strings.HasPrefix(v.Version, "3."),
			"版本应该以2.或3.开头: %s", v.Version)

		// 不应该包含预发布标记
		lowercaseVersion := strings.ToLower(v.Version)
		assert.False(t, strings.Contains(lowercaseVersion, "beta"), "版本不应包含beta: %s", v.Version)
		assert.False(t, strings.Contains(lowercaseVersion, "alpha"), "版本不应包含alpha: %s", v.Version)
		assert.False(t, strings.Contains(lowercaseVersion, "rc"), "版本不应包含rc: %s", v.Version)
		assert.False(t, strings.Contains(lowercaseVersion, "snapshot"), "版本不应包含snapshot: %s", v.Version)
	}
}

// 测试使用模拟服务器的GetVersionsWithMetadata功能
func TestGetVersionsWithMetadataMock(t *testing.T) {
	// 设置模拟服务器
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/solrsearch/select" && r.Method == http.MethodGet {
			// 返回版本列表
			versions := []*response.Version{
				{Version: "1.0.0", GroupId: "org.example", ArtifactId: "test-artifact"},
				{Version: "1.1.0", GroupId: "org.example", ArtifactId: "test-artifact"},
			}
			mockSearchResponse(w, versions, len(versions))
			return
		}

		if strings.HasPrefix(r.URL.Path, "/v1/versions/") {
			// 返回版本详细信息
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) >= 5 { // 确保有足够的路径段
				version := parts[4]
				versionInfo := response.VersionInfo{
					GroupId:     "org.example",
					ArtifactId:  "test-artifact",
					Version:     version,
					LastUpdated: fmt.Sprintf("2023-01-0%d:00:00.000Z", len(version)), // 生成一个不同的时间戳
					Packaging:   "jar",
				}
				json.NewEncoder(w).Encode(versionInfo)
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
	})

	// 测试GetVersionsWithMetadata
	versionsWithMeta, err := client.GetVersionsWithMetadata(context.Background(), "org.example", "test-artifact")
	assert.NoError(t, err)
	assert.NotNil(t, versionsWithMeta)
	assert.Equal(t, 2, len(versionsWithMeta), "应该返回2个带元数据的版本")

	// 检查版本和元数据是否匹配
	for _, v := range versionsWithMeta {
		assert.Equal(t, v.Version.Version, v.VersionInfo.Version, "版本号应匹配")
		assert.Equal(t, "org.example", v.VersionInfo.GroupId, "GroupId应匹配")
		assert.Equal(t, "test-artifact", v.VersionInfo.ArtifactId, "ArtifactId应匹配")
		assert.NotEmpty(t, v.VersionInfo.LastUpdated, "LastUpdated不应为空")
	}
}

// 测试使用模拟服务器的CompareVersions功能
func TestCompareVersionsMock(t *testing.T) {
	// 设置模拟服务器
	_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v1/versions/") {
			// 从URL路径解析版本号
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) < 5 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			groupId := parts[2]
			artifactId := parts[3]
			version := parts[4]

			// 根据不同版本返回不同的响应
			var lastUpdated string
			switch version {
			case "1.0.0":
				lastUpdated = "2022-01-01T00:00:00.000Z"
			case "2.0.0":
				lastUpdated = "2023-01-01T00:00:00.000Z"
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}

			versionInfo := response.VersionInfo{
				GroupId:     groupId,
				ArtifactId:  artifactId,
				Version:     version,
				LastUpdated: lastUpdated,
				Packaging:   "jar",
			}
			json.NewEncoder(w).Encode(versionInfo)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// 测试版本比较
	comparison, err := client.CompareVersions(context.Background(), "test.group", "test.artifact", "1.0.0", "2.0.0")
	assert.NoError(t, err)
	assert.NotNil(t, comparison)

	// 验证比较结果
	assert.Equal(t, "1.0.0", comparison.Version1)
	assert.Equal(t, "2.0.0", comparison.Version2)
	assert.Equal(t, "2022-01-01T00:00:00.000Z", comparison.V1Timestamp)
	assert.Equal(t, "2023-01-01T00:00:00.000Z", comparison.V2Timestamp)

	// 测试比较相同版本
	comparison, err = client.CompareVersions(context.Background(), "test.group", "test.artifact", "1.0.0", "1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, comparison.V1Timestamp, comparison.V2Timestamp)

	// 测试比较不存在的版本
	_, err = client.CompareVersions(context.Background(), "test.group", "test.artifact", "1.0.0", "999.0.0")
	assert.Error(t, err, "比较不存在的版本应该返回错误")
}

// 测试GetVersionInfo错误处理
func TestGetVersionInfoErrorMock(t *testing.T) {
	t.Run("版本不存在", func(t *testing.T) {
		// 设置模拟服务器 - 返回404
		_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/v1/versions/") {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error": "Version not found"}`))
			}
		})

		// 测试获取不存在的版本
		info, err := client.GetVersionInfo(context.Background(), "org.nonexistent", "nonexistent-artifact", "1.0.0")
		assert.Error(t, err, "应该返回错误")
		assert.Nil(t, info, "不应返回版本信息")
		assert.Contains(t, err.Error(), "404", "错误应包含404状态码")
	})

	t.Run("无效响应", func(t *testing.T) {
		// 设置模拟服务器 - 返回无效JSON
		_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/v1/versions/") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"invalid-json`)) // 无效的JSON
			}
		})

		// 测试解析无效响应
		info, err := client.GetVersionInfo(context.Background(), "org.example", "test-artifact", "1.0.0")
		assert.Error(t, err, "应该返回错误")
		assert.Nil(t, info, "不应返回版本信息")
		assert.Contains(t, err.Error(), "JSON", "错误应包含JSON解析失败提示")
	})

	t.Run("服务器错误", func(t *testing.T) {
		// 设置模拟服务器 - 返回500错误
		_, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/v1/versions/") {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Internal server error"}`))
			}
		})

		// 测试服务器错误
		info, err := client.GetVersionInfo(context.Background(), "org.example", "test-artifact", "1.0.0")
		assert.Error(t, err, "应该返回错误")
		assert.Nil(t, info, "不应返回版本信息")
		assert.Contains(t, err.Error(), "500", "错误应包含500状态码")
	})
}
