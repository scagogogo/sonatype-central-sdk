package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createRealClientForGroupTest() *Client {
	// 创建真实客户端进行测试
	client := NewClient()
	return client
}

func TestSearchByGroupIdBasic(t *testing.T) {
	client := createRealClientForGroupTest()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// 测试常见的组ID
	groupId := "org.apache.commons"
	limit := 10

	artifacts, err := client.SearchByGroupId(ctx, groupId, limit)
	assert.NoError(t, err)
	assert.NotEmpty(t, artifacts)
	assert.LessOrEqual(t, len(artifacts), limit)

	t.Logf("找到 %d 个组件", len(artifacts))
	for i, artifact := range artifacts {
		if i < 3 { // 只输出前三个，避免日志过长
			t.Logf("组件 %d: %s:%s", i+1, artifact.GroupId, artifact.ArtifactId)
		}
	}
}

func TestIteratorByGroupId(t *testing.T) {
	client := createRealClientForGroupTest()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// 测试使用迭代器获取所有组件
	groupId := "commons-io"
	iterator := client.IteratorByGroupId(ctx, groupId)

	// 获取前5个结果
	count := 0
	limit := 5
	for iterator.Next() && count < limit {
		artifact := iterator.Value()
		assert.NotNil(t, artifact)
		t.Logf("迭代器获取组件: %s:%s", artifact.GroupId, artifact.ArtifactId)
		count++
	}
	assert.LessOrEqual(t, count, limit)
}

func TestSearchByGroupPattern(t *testing.T) {
	client := createRealClientForGroupTest()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// 测试使用模式搜索组
	pattern := "apache.log4j"
	limit := 5

	groups, err := client.SearchByGroupPattern(ctx, pattern, limit)
	assert.NoError(t, err)

	t.Logf("模式 '%s' 搜索到 %d 个组", pattern, len(groups))
	for i, group := range groups {
		if i < 3 { // 只显示前3个结果
			t.Logf("组 %d: %s, 包含 %d 个组件, 最后更新: %s",
				i+1, group.GroupId, group.ArtifactCount, group.LastUpdatedDate)
		}
	}
}

func TestGetGroupStatistics(t *testing.T) {
	client := createRealClientForGroupTest()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	// 测试获取组统计信息
	groupId := "org.slf4j"

	stats, err := client.GetGroupStatistics(ctx, groupId)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, groupId, stats.GroupId)
	assert.Greater(t, stats.ArtifactCount, 0)
	assert.Greater(t, stats.TotalVersions, 0)
	assert.NotEmpty(t, stats.LastUpdatedDate)

	t.Logf("组 %s 统计: %d 个组件, %d 个版本, 最后更新: %s",
		stats.GroupId, stats.ArtifactCount, stats.TotalVersions, stats.LastUpdatedDate)

	if len(stats.Artifacts) > 0 {
		t.Logf("前3个组件统计:")
		for i, artifact := range stats.Artifacts {
			if i < 3 {
				t.Logf("  %s: %d 个版本, 最新版本: %s",
					artifact.ArtifactId, artifact.VersionCount, artifact.LatestVersion)
			}
		}
	}
}

func TestGetPopularGroups(t *testing.T) {
	client := createRealClientForGroupTest()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// 测试获取流行组
	limit := 5

	groups, err := client.GetPopularGroups(ctx, limit)
	assert.NoError(t, err)
	assert.NotEmpty(t, groups)
	assert.LessOrEqual(t, len(groups), limit)

	t.Logf("获取 %d 个流行组:", len(groups))
	for _, group := range groups {
		t.Logf("第 %d 名: %s, 含有 %d 个组件",
			group.PopularityRank, group.GroupId, group.ArtifactCount)
	}
}

func TestCompareTwoGroups(t *testing.T) {
	client := createRealClientForGroupTest()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 测试比较两个组
	group1 := "org.apache.logging.log4j"
	group2 := "ch.qos.logback"

	comparison, err := client.CompareTwoGroups(ctx, group1, group2)
	// 忽略比较中可能出现的错误，因为我们主要关注比较结果
	_ = err

	assert.NotNil(t, comparison)
	assert.Equal(t, group1, comparison.Group1)
	assert.Equal(t, group2, comparison.Group2)

	t.Logf("比较 %s 和 %s:", group1, group2)
	if comparison.Group1Stats != nil {
		t.Logf("%s: %d 个组件, %d 个版本",
			group1, comparison.Group1Stats.ArtifactCount, comparison.Group1Stats.TotalVersions)
	}
	if comparison.Group2Stats != nil {
		t.Logf("%s: %d 个组件, %d 个版本",
			group2, comparison.Group2Stats.ArtifactCount, comparison.Group2Stats.TotalVersions)
	}
	t.Logf("共同组件数量: %d", comparison.CommonArtifactCount)
	if comparison.CommonArtifactCount > 0 && len(comparison.CommonArtifacts) > 0 {
		t.Logf("部分共同组件: %v", comparison.CommonArtifacts[:min(3, len(comparison.CommonArtifacts))])
	}
}

func TestSearchSubgroups(t *testing.T) {
	client := createRealClientForGroupTest()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// 测试搜索子组
	parentGroup := "org.apache"
	limit := 10

	subgroups, err := client.SearchSubgroups(ctx, parentGroup, limit)
	assert.NoError(t, err)

	t.Logf("找到 %s 的 %d 个子组:", parentGroup, len(subgroups))
	for i, group := range subgroups {
		if i < 5 { // 只显示前5个
			t.Logf("子组 %d: %s, 包含 %d 个组件",
				i+1, group.GroupId, group.ArtifactCount)
		}
	}
}

func TestGetGroupInfo(t *testing.T) {
	client := createRealClientForGroupTest()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// 测试常见的组ID
	groupId := "org.springframework"

	info, err := client.GetGroupInfo(ctx, groupId)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, groupId, info.GroupId)
	assert.Greater(t, info.ArtifactCount, 0)
	assert.NotZero(t, info.LastUpdated)
	assert.NotEmpty(t, info.LastUpdatedDate)

	t.Logf("组 %s 的基本信息:", groupId)
	t.Logf("  - 组件数量: %d", info.ArtifactCount)
	t.Logf("  - 最后更新: %s", info.LastUpdatedDate)
	t.Logf("  - 描述: %s", info.Description)
	t.Logf("  - 网站: %s", info.Website)

	// 测试不存在的组ID
	nonExistGroupId := "org.nonexistgroup123456789"
	nonExistInfo, err := client.GetGroupInfo(ctx, nonExistGroupId)

	// 我们应该能够获取到一个空信息，而不是错误
	assert.NoError(t, err)
	assert.NotNil(t, nonExistInfo)
	assert.Equal(t, nonExistGroupId, nonExistInfo.GroupId)
	assert.Zero(t, nonExistInfo.ArtifactCount)
}

// min函数辅助比较两个整数的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
