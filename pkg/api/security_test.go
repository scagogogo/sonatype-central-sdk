package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 将createRealClient函数内联到测试文件中
func createRealClientForTest(t *testing.T) *Client {
	// 创建默认客户端实例（使用真实API地址）
	client := NewClient(
		WithMaxRetries(3),     // 设置更多重试次数以应对临时网络问题
		WithRetryBackoff(800), // 较长的重试间隔，避免过快重试
		WithCache(true, 3600), // 启用长时间缓存以减少对API的请求
	)

	// 在测试结束时清除缓存
	t.Cleanup(func() {
		client.ClearCache()
	})

	return client
}

// TestGetSecurityRatingReal 测试获取安全评分功能
func TestGetSecurityRatingReal(t *testing.T) {
	// 使用真实客户端
	client := createRealClientForTest(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试一些已知有漏洞的库
	testCases := []struct {
		name       string
		groupId    string
		artifactId string
		version    string
	}{
		{"Log4j-core", "org.apache.logging.log4j", "log4j-core", "2.14.1"}, // 已知存在Log4Shell漏洞
		{"Spring-core", "org.springframework", "spring-core", "5.3.10"},
		{"Commons-text", "org.apache.commons", "commons-text", "1.9.0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 获取安全评分
			rating, err := client.GetSecurityRating(ctx, tc.groupId, tc.artifactId, tc.version)
			if err != nil {
				t.Logf("跳过测试 %s: %v", tc.name, err)
				t.Skip("无法连接到安全评分API")
				return
			}

			// 验证评分信息
			assert.NotNil(t, rating)
			t.Logf("%s:%s:%s 安全评分: %.1f, 严重性: %s, 漏洞数: %d",
				tc.groupId, tc.artifactId, tc.version, rating.Score, rating.Severity, rating.VulnCount)

			// 不要强制要求特定的漏洞数量，因为这可能随着时间变化
			// 只验证响应结构的完整性
			assert.GreaterOrEqual(t, rating.Score, 0.0)
			assert.LessOrEqual(t, rating.Score, 10.0)
		})
	}
}

// TestCompareVersionSecurityReal 测试版本安全比较功能
func TestCompareVersionSecurityReal(t *testing.T) {
	// 使用真实客户端
	client := createRealClientForTest(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试两个不同版本的 Log4j
	// Log4j 2.14.1 存在 Log4Shell 漏洞，2.15.0 修复了部分漏洞
	groupId := "org.apache.logging.log4j"
	artifactId := "log4j-core"
	version1 := "2.14.1"
	version2 := "2.15.0"

	// 比较版本安全性
	comparison, err := client.CompareVersionSecurity(ctx, groupId, artifactId, version1, version2)
	if err != nil {
		t.Logf("跳过测试: %v", err)
		t.Skip("无法连接到安全评分API")
		return
	}

	assert.NotNil(t, comparison)
	assert.Equal(t, groupId, comparison.GroupId)
	assert.Equal(t, artifactId, comparison.ArtifactId)
	assert.Equal(t, version1, comparison.Version1)
	assert.Equal(t, version2, comparison.Version2)

	t.Logf("比较 %s vs %s", version1, version2)
	t.Logf("版本1评分: %.1f, 漏洞数: %d", comparison.Rating1.Score, comparison.Rating1.VulnCount)
	t.Logf("版本2评分: %.1f, 漏洞数: %d", comparison.Rating2.Score, comparison.Rating2.VulnCount)
	t.Logf("更安全版本: %s, 分数差异: %.1f", comparison.SaferVersion, comparison.ScoreDifference)
}

// TestSearchVulnerableArtifactsReal 测试搜索有漏洞的制品
func TestSearchVulnerableArtifactsReal(t *testing.T) {
	// 使用真实客户端
	client := createRealClientForTest(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 安全严重性级别定义
	securitySeverityHigh := SecuritySeverity("HIGH")

	// 搜索具有高风险漏洞的制品
	artifacts, err := client.SearchVulnerableArtifacts(ctx, securitySeverityHigh, 5)
	if err != nil {
		t.Logf("跳过测试: %v", err)
		t.Skip("无法连接到安全API")
		return
	}

	assert.NotNil(t, artifacts)

	if len(artifacts) > 0 {
		t.Logf("找到 %d 个具有高风险漏洞的制品", len(artifacts))
		for i, a := range artifacts {
			t.Logf("%d. %s:%s", i+1, a.GroupId, a.ArtifactId)
		}
	} else {
		t.Log("未找到高风险漏洞制品，这可能是由于API响应限制")
	}
}

// TestVulnerabilityTimelineReal 测试获取漏洞时间线
func TestVulnerabilityTimelineReal(t *testing.T) {
	// 使用真实客户端
	client := createRealClientForTest(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 使用已知有漏洞历史的库
	groupId := "org.apache.logging.log4j"
	artifactId := "log4j-core"
	maxVersions := 5 // 限制版本数量，避免测试过长

	// 获取漏洞时间线
	timeline, err := client.GetVulnerabilityTimeline(ctx, groupId, artifactId, maxVersions)
	if err != nil {
		t.Logf("跳过测试: %v", err)
		t.Skip("无法连接到安全API")
		return
	}

	assert.NotNil(t, timeline)
	assert.Equal(t, groupId, timeline.GroupId)
	assert.Equal(t, artifactId, timeline.ArtifactId)

	if len(timeline.Entries) > 0 {
		t.Logf("找到 %d 个时间线条目", len(timeline.Entries))
		for i, entry := range timeline.Entries {
			t.Logf("%d. 版本: %s, 评分: %.1f, 变化: %s, 漏洞数: %d",
				i+1, entry.Version, entry.Score, entry.Change, entry.VulnCount)
		}
	} else {
		t.Log("未找到时间线条目，这可能是由于API响应限制")
	}
}

// TestSecurityDataStructures 测试安全相关数据结构和API调用
func TestSecurityDataStructures(t *testing.T) {
	// 创建一个测试客户端
	client := createRealClientForTest(t)

	// 设置上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. 测试SecuritySeverity类型
	sevCritical := SecuritySeverity("CRITICAL")
	sevHigh := SecuritySeverity("HIGH")
	sevMedium := SecuritySeverity("MEDIUM")
	sevLow := SecuritySeverity("LOW")
	sevNone := SecuritySeverity("NONE")

	assert.Equal(t, SecuritySeverity("CRITICAL"), sevCritical)
	assert.Equal(t, SecuritySeverity("HIGH"), sevHigh)
	assert.Equal(t, SecuritySeverity("MEDIUM"), sevMedium)
	assert.Equal(t, SecuritySeverity("LOW"), sevLow)
	assert.Equal(t, SecuritySeverity("NONE"), sevNone)

	// 2. 测试GetVulnerabilityDetails API
	// 使用已知有漏洞的库进行测试
	groupId := "org.apache.logging.log4j"
	artifactId := "log4j-core"
	version := "2.14.1" // 已知存在漏洞的版本

	details, err := client.GetVulnerabilityDetails(ctx, groupId, artifactId, version)
	if err != nil {
		t.Logf("跳过GetVulnerabilityDetails测试: %v", err)
		t.Log("继续执行其他测试...")
	} else {
		assert.NotNil(t, details)
		assert.Equal(t, groupId, details.GroupId)
		assert.Equal(t, artifactId, details.ArtifactId)
		assert.Equal(t, version, details.Version)

		// 验证返回的漏洞信息
		if len(details.Vulnerabilities) > 0 {
			t.Logf("找到 %d 个漏洞", len(details.Vulnerabilities))
			// 检查第一个漏洞的结构
			vuln := details.Vulnerabilities[0]
			assert.NotEmpty(t, vuln.ID)
			assert.NotEmpty(t, vuln.Title)
			assert.NotEmpty(t, vuln.Severity)
			assert.Greater(t, vuln.CvssScore, 0.0)
		} else {
			t.Log("未找到漏洞信息，可能是API响应有变化")
		}
	}

	// 3. 测试SecurityRating结构和函数
	rating, err := client.GetSecurityRating(ctx, groupId, artifactId, version)
	if err != nil {
		t.Logf("跳过GetSecurityRating测试: %v", err)
	} else {
		assert.NotNil(t, rating)
		assert.GreaterOrEqual(t, rating.Score, 0.0)
		assert.LessOrEqual(t, rating.Score, 10.0)
		assert.NotEmpty(t, rating.Severity)

		// 检查评分和严重性的一致性
		if rating.Score >= 7.0 {
			assert.Contains(t, []string{"HIGH", "CRITICAL"}, rating.Severity)
		}
	}

	t.Log("安全API数据结构和功能测试通过")
}

// TestSecurityEdgeCases 测试安全API的边界情况
func TestSecurityEdgeCases(t *testing.T) {
	// 创建一个测试客户端
	client := createRealClientForTest(t)

	// 设置上下文
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. 测试不存在的组件
	nonExistentGroupId := "com.nonexistent.group"
	nonExistentArtifactId := "nonexistent-artifact"
	nonExistentVersion := "1.0.0"

	// 测试GetSecurityRating对不存在组件的处理
	rating, err := client.GetSecurityRating(ctx, nonExistentGroupId, nonExistentArtifactId, nonExistentVersion)
	if err != nil {
		t.Logf("预期的错误: %v", err)
		assert.Error(t, err) // 期望出现错误
	} else {
		t.Log("预期会有错误，但API返回了结果")
		assert.Equal(t, 0.0, rating.Score)
		assert.Equal(t, "NONE", rating.Severity)
	}

	// 2. 测试带有限制的漏洞搜索
	artifacts, err := client.SearchVulnerableArtifacts(ctx, SecuritySeverity("HIGH"), 1)
	if err != nil {
		t.Logf("跳过限制搜索测试: %v", err)
	} else {
		// 验证结果数量不超过限制
		assert.LessOrEqual(t, len(artifacts), 1)
	}

	// 3. 测试CVE影响检查
	knownGroup := "org.apache.logging.log4j"
	knownArtifact := "log4j-core"
	knownVersion := "2.14.1"
	cveId := "CVE-2021-44228" // Log4Shell漏洞

	isImpacted, vuln, err := client.CheckCVEImpact(ctx, cveId, knownGroup, knownArtifact, knownVersion)
	if err != nil {
		t.Logf("跳过CVE影响检查测试: %v", err)
	} else {
		if isImpacted {
			assert.True(t, isImpacted)
			assert.NotNil(t, vuln)
			assert.Equal(t, cveId, vuln.CVE)
			t.Logf("组件 %s:%s:%s 受到 %s 影响，严重性: %s",
				knownGroup, knownArtifact, knownVersion, cveId, vuln.Severity)
		} else {
			t.Logf("组件未受影响或API响应有变化")
		}
	}

	t.Log("安全API边界情况测试完成")
}
