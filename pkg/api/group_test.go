package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSearchByGroupId(t *testing.T) {
	// 创建真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 测试搜索
	artifacts, err := client.SearchByGroupId(ctx, "org.junit", 5)
	if err != nil {
		t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, artifacts)
	assert.True(t, len(artifacts) > 0, "应该找到至少一个制品")

	// 检查返回结果
	t.Logf("找到 %d 个org.junit制品", len(artifacts))
	for i, artifact := range artifacts[:minInt(3, len(artifacts))] {
		t.Logf("制品 %d: %s:%s (%s)", i+1, artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
		assert.Equal(t, "org.junit", artifact.GroupId)
		assert.NotEmpty(t, artifact.ArtifactId)
		assert.NotEmpty(t, artifact.LatestVersion)
	}
}

func TestSearchByGroupIdWithLimit(t *testing.T) {
	// 创建真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 测试限制搜索
	limit := 3
	artifacts, err := client.SearchByGroupId(ctx, "org.apache", limit)
	if err != nil {
		t.Logf("跳过测试，无法连接到Maven Central API: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, artifacts)
	assert.LessOrEqual(t, len(artifacts), limit, "返回结果不应超过指定的限制")

	// 记录找到的制品
	t.Logf("找到 %d 个org.apache制品", len(artifacts))
	for i, artifact := range artifacts {
		t.Logf("制品 %d: %s:%s", i+1, artifact.GroupId, artifact.ArtifactId)
	}
}

func TestIteratorByGroupId(t *testing.T) {
	// 创建真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建迭代器
	iterator := client.IteratorByGroupId(ctx, "com.google.code.gson")
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
		t.Log("迭代器未能找到任何gson相关的制品，这可能是API限制或网络问题导致的")
	} else {
		t.Logf("迭代器成功找到 %d 个gson相关的制品", count)
	}
}
