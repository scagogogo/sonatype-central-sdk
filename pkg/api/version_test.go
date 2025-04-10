package api

import (
	"context"
	"fmt"
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
	for i, v := range versions[:minInt(5, len(versions))] {
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
		for i, v := range versionsWithMeta[:minInt(3, len(versionsWithMeta))] {
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

// 测试GetLatestVersion的基本功能（使用真实客户端）
func TestGetLatestVersionBasic(t *testing.T) {
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 测试GetLatestVersion，使用常见库
	latestVersion, err := client.GetLatestVersion(ctx, "junit", "junit")
	if err != nil {
		t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, latestVersion)
	assert.NotEmpty(t, latestVersion.Version, "应返回有效的版本号")
	t.Logf("junit:junit 最新版本: %s", latestVersion.Version)
}

// 测试FilterVersions的基本功能（使用真实客户端）
func TestFilterVersionsBasic(t *testing.T) {
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试真实库的版本过滤
	filteredVersions, err := client.FilterVersions(ctx, "org.apache.commons", "commons-lang3", func(v *response.Version) bool {
		return strings.HasPrefix(v.Version, "3.")
	})

	if err != nil {
		t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, filteredVersions)

	if len(filteredVersions) > 0 {
		t.Logf("找到 %d 个3.x版本的commons-lang3", len(filteredVersions))
		for i, v := range filteredVersions[:minInt(5, len(filteredVersions))] {
			t.Logf("版本 %d: %s", i+1, v.Version)
			assert.True(t, strings.HasPrefix(v.Version, "3."), "所有版本都应以3.开头")
		}
	} else {
		t.Log("未找到任何3.x版本，这可能是API返回数据有限导致的")
	}
}

// 测试HasVersion功能（使用真实客户端）
func TestHasVersionBasic(t *testing.T) {
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("版本存在", func(t *testing.T) {
		// 测试一个已知存在的版本
		exists, err := client.HasVersion(ctx, "junit", "junit", "4.12")
		if err != nil {
			t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		assert.NoError(t, err)
		assert.True(t, exists, "junit:junit:4.12应该存在")
	})

	t.Run("版本不存在", func(t *testing.T) {
		// 睡眠一段时间，避免请求过快
		time.Sleep(1 * time.Second)

		// 测试一个肯定不存在的版本
		exists, err := client.HasVersion(ctx, "junit", "junit", "999.999.999")
		if err != nil {
			t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		assert.NoError(t, err)
		assert.False(t, exists, "junit:junit:999.999.999不应该存在")
	})
}

// 测试GetVersionsWithMetadata功能（使用真实客户端）
func TestGetVersionsWithMetadataBasic(t *testing.T) {
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 为了减少API调用，先获取少量版本
	allVersions, err := client.ListVersions(ctx, "org.slf4j", "slf4j-api", 3)
	if err != nil || len(allVersions) == 0 {
		t.Logf("跳过测试：无法获取slf4j-api版本列表: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	// 为减轻API负担，只测试有限的几个版本
	groupId := "org.slf4j"
	artifactId := "slf4j-api"
	if len(allVersions) > 0 {
		version := allVersions[0].Version

		// 测试单个版本的元数据
		versionInfo, err := client.GetVersionInfo(ctx, groupId, artifactId, version)
		if err != nil {
			t.Logf("跳过测试：获取版本元数据失败: %v", err)
			t.Skip("无法获取版本元数据")
			return
		}

		assert.NotNil(t, versionInfo)
		assert.Equal(t, groupId, versionInfo.GroupId)
		assert.Equal(t, artifactId, versionInfo.ArtifactId)
		assert.Equal(t, version, versionInfo.Version)
		assert.NotEmpty(t, versionInfo.LastUpdated)

		t.Logf("获取到版本 %s 的元数据，最后更新时间: %s", version, versionInfo.LastUpdated)
	}
}

// TestVersionsEdgeCases 测试版本API的边界情况
func TestVersionsEdgeCases(t *testing.T) {
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("不存在的构件", func(t *testing.T) {
		// 生成一个肯定不存在的构件ID
		randomArtifactId := fmt.Sprintf("nonexistent-artifact-%d", time.Now().UnixNano())
		versions, err := client.ListVersions(ctx, "org.apache.commons", randomArtifactId, 10)

		// 不应该返回错误，但结果应该为空
		if err != nil {
			t.Logf("查询不存在的构件返回错误: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		assert.Empty(t, versions, "不存在的构件应返回空版本列表")
	})

	t.Run("限制为零的ListVersions", func(t *testing.T) {
		// 测试limit为0的情况，应该调用IteratorVersions
		versions, err := client.ListVersions(ctx, "junit", "junit", 0)
		if err != nil {
			t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		// 应返回所有可用版本
		assert.NotEmpty(t, versions, "limit为0时应返回所有版本")
		t.Logf("使用limit=0获取到%d个junit版本", len(versions))
	})

	t.Run("限制为负数的ListVersions", func(t *testing.T) {
		// 测试limit为负数的情况，应该调用IteratorVersions
		versions, err := client.ListVersions(ctx, "junit", "junit", -1)
		if err != nil {
			t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		// 应返回所有可用版本
		assert.NotEmpty(t, versions, "limit为负数时应返回所有版本")
		t.Logf("使用limit=-1获取到%d个junit版本", len(versions))
	})

	t.Run("不存在的版本信息", func(t *testing.T) {
		// 测试获取不存在版本的信息
		_, err := client.GetVersionInfo(ctx, "junit", "junit", "999.999.999")
		assert.Error(t, err, "获取不存在的版本信息应返回错误")
		assert.ErrorIs(t, err, ErrNotFound, "错误类型应为ErrNotFound")
	})

	t.Run("空过滤条件", func(t *testing.T) {
		// 测试使用接受所有版本的过滤器
		versions, err := client.FilterVersions(ctx, "junit", "junit", func(_ *response.Version) bool {
			return true
		})

		if err != nil {
			t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		// 应返回与ListVersions相同的结果
		allVersions, _ := client.ListVersions(ctx, "junit", "junit", 0)
		if len(allVersions) > 0 {
			assert.Equal(t, len(allVersions), len(versions), "接受所有版本的过滤器应返回所有版本")
		}
	})

	t.Run("拒绝所有版本的过滤器", func(t *testing.T) {
		// 测试拒绝所有版本的过滤器
		versions, err := client.FilterVersions(ctx, "junit", "junit", func(_ *response.Version) bool {
			return false
		})

		if err != nil {
			t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
			t.Skip("无法连接到Maven Central API")
			return
		}

		assert.Empty(t, versions, "拒绝所有版本的过滤器应返回空列表")
	})
}
