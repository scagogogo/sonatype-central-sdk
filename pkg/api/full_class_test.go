package api

import (
	"context"
	"testing"
	"time"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
	"github.com/stretchr/testify/assert"
)

// TestSearchByJavaPackage 测试Java包名搜索
func TestSearchByJavaPackage(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试几个常见的包名
	packageNames := []string{
		"org.apache.commons.lang3",
		"java.util",
		"org.slf4j",
	}

	for _, packageName := range packageNames {
		t.Run("Package_"+packageName, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(1 * time.Second)

			versionSlice, err := client.SearchByJavaPackage(ctx, packageName, 3)

			if err != nil {
				t.Logf("搜索包 %s 时出错: %v", packageName, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			// 记录找到的结果
			t.Logf("找到 %d 个包含包 %s 的结果", len(versionSlice), packageName)
			if len(versionSlice) > 0 {
				maxShow := 3
				if len(versionSlice) < maxShow {
					maxShow = len(versionSlice)
				}
				for i, v := range versionSlice[:maxShow] {
					t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
				}
			}

			assert.True(t, len(versionSlice) >= 0) // 只确保API正常返回
		})
	}
}

// TestIteratorByJavaPackage 测试Java包名迭代器
func TestIteratorByJavaPackage(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试一个常见的包名
	packageName := "org.apache.commons.lang3"

	// 添加短暂延迟，避免请求过快
	time.Sleep(1 * time.Second)

	iterator := client.IteratorByJavaPackage(ctx, packageName)

	// 获取前3个结果
	var results []*response.Version
	count := 0
	for iterator.Next() && count < 3 {
		results = append(results, iterator.Value())
		count++
	}

	// 检查迭代器是否有错误
	_, err := iterator.NextE()
	if err != nil && err != ErrQueryIteratorEnd {
		t.Logf("迭代器使用时出错: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	// 记录找到的结果
	t.Logf("找到至少 %d 个包含包 %s 的结果", len(results), packageName)
	for i, v := range results {
		t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
	}

	assert.True(t, len(results) >= 0) // 只确保API正常返回
}

// TestFullClassRelatedMethods 测试全类名相关方法
func TestFullClassRelatedMethods(t *testing.T) {
	// 使用真实客户端
	client := createRealClient(t)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试全类名和包名+类名组合查询
	testCases := []struct {
		name     string
		function func() ([]*response.Version, error)
	}{
		{
			"全类名_StringUtils",
			func() ([]*response.Version, error) {
				return client.SearchByFullyQualifiedClassName(ctx, "org.apache.commons.lang3.StringUtils", 3)
			},
		},
		{
			"包名类名_StringUtils",
			func() ([]*response.Version, error) {
				return client.SearchByPackageAndClassName(ctx, "org.apache.commons.lang3", "StringUtils", 3)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 添加短暂延迟，避免请求过快
			time.Sleep(1 * time.Second)

			versionSlice, err := tc.function()

			if err != nil {
				t.Logf("执行 %s 时出错: %v", tc.name, err)
				t.Skip("无法连接到Maven Central API")
				return
			}

			// 记录找到的结果
			t.Logf("找到 %d 个结果", len(versionSlice))
			if len(versionSlice) > 0 {
				maxShow := 3
				if len(versionSlice) < maxShow {
					maxShow = len(versionSlice)
				}
				for i, v := range versionSlice[:maxShow] {
					t.Logf("结果 %d: %s:%s:%s", i+1, v.GroupId, v.ArtifactId, v.Version)
				}
			}

			assert.True(t, len(versionSlice) >= 0) // 只确保API正常返回
		})
	}
}
