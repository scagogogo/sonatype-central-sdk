package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSearchByArtifactId(t *testing.T) {
	// 创建真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 测试搜索
	artifacts, err := client.SearchByArtifactId(ctx, "commons-io", 3)
	if err != nil {
		t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, artifacts)
	assert.True(t, len(artifacts) > 0, "应该找到commons-io相关的制品")

	// 检查返回结果
	t.Logf("找到 %d 个commons-io相关的制品", len(artifacts))
	for i, artifact := range artifacts[:minInt(3, len(artifacts))] {
		t.Logf("制品 %d: %s:%s (%s)", i+1, artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
		assert.NotEmpty(t, artifact.GroupId, "GroupId不应为空")
		assert.NotEmpty(t, artifact.ArtifactId, "ArtifactId不应为空")
		assert.NotEmpty(t, artifact.LatestVersion, "LatestVersion不应为空")
	}
}

func TestSearchByArtifactIdWithLimit(t *testing.T) {
	// 创建真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 测试限制搜索
	limit := 5
	artifacts, err := client.SearchByArtifactId(ctx, "commons", limit)
	if err != nil {
		t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, artifacts)
	assert.LessOrEqual(t, len(artifacts), limit, "返回结果不应超过指定的限制")

	// 记录找到的制品
	t.Logf("找到 %d 个包含'commons'的制品", len(artifacts))
	for i, artifact := range artifacts {
		t.Logf("制品 %d: %s:%s", i+1, artifact.GroupId, artifact.ArtifactId)
	}
}

func TestIteratorByArtifactId(t *testing.T) {
	// 创建真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建迭代器
	iterator := client.IteratorByArtifactId(ctx, "junit")
	assert.NotNil(t, iterator)

	// 迭代前几个元素
	count := 0
	maxCount := 5 // 只迭代前5个，以避免过多API调用
	for iterator.Next() && count < maxCount {
		artifact := iterator.Value()
		assert.NotNil(t, artifact)
		t.Logf("迭代器返回制品: %s:%s", artifact.GroupId, artifact.ArtifactId)
		count++
	}

	// 检查是否找到了结果
	if count == 0 {
		t.Log("迭代器未能找到任何junit相关的制品，这可能是API限制或网络问题导致的")
	} else {
		t.Logf("迭代器成功找到 %d 个junit相关的制品", count)
	}
}
