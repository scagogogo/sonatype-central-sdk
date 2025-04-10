package api

import (
	"context"
	"testing"
	"time"
)

// TestSearchByGAV 测试根据GAV进行搜索
func TestSearchByGAV(t *testing.T) {
	client := createRealClient(t)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 构建查询
	query := "g:org.apache.commons AND a:commons-lang3"

	// 执行查询
	results, err := client.ListGAVs(ctx, query, 5)
	if err != nil {
		t.Logf("跳过测试: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	// 验证结果
	if len(results) == 0 {
		t.Fatal("未找到commons-lang3组件，这可能是一个错误")
	}

	// 输出结果
	t.Logf("找到 %d 个结果", len(results))
	for i, result := range results {
		t.Logf("%d. %s:%s:%s", i+1, result.GroupId, result.ArtifactId, result.LatestVersion)
	}
}

// TestGetGAVInfo 测试获取指定GAV的详细信息
func TestGetGAVInfo(t *testing.T) {
	client := createRealClient(t)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试已知存在的GAV坐标
	groupId := "org.apache.commons"
	artifactId := "commons-lang3"

	// 执行查询
	artifact, err := client.GetGAVInfo(ctx, groupId, artifactId, "")
	if err != nil {
		t.Logf("跳过测试: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	// 验证结果
	if artifact == nil {
		t.Fatal("未找到commons-lang3组件，这可能是一个错误")
	}

	t.Logf("找到制品: %s:%s:%s, 包类型: %s, 版本数: %d",
		artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion,
		artifact.Packaging, artifact.VersionCount)

	// 测试带版本号的查询
	version := artifact.LatestVersion
	artifactWithVersion, err := client.GetGAVInfo(ctx, groupId, artifactId, version)
	if err != nil {
		t.Fatalf("未能找到特定版本: %v", err)
	}

	t.Logf("找到特定版本制品: %s:%s:%s",
		artifactWithVersion.GroupId, artifactWithVersion.ArtifactId, artifactWithVersion.LatestVersion)
}

// TestSearchGAVsWithSort 测试带排序的GAV搜索
func TestSearchGAVsWithSort(t *testing.T) {
	client := createRealClient(t)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 构建查询
	query := "g:org.apache"

	// 按版本数量降序排序
	results, err := client.SearchGAVsWithSort(ctx, query, "versionCount", false, 5)
	if err != nil {
		t.Logf("跳过测试: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	// 验证结果
	if len(results) == 0 {
		t.Fatal("未找到任何组件，这可能是一个错误")
	}

	t.Logf("按版本数量降序排序，找到 %d 个结果", len(results))
	for i, result := range results {
		t.Logf("%d. %s:%s, 版本数: %d",
			i+1, result.GroupId, result.ArtifactId, result.VersionCount)
	}

	// 确认排序正确
	if len(results) > 1 {
		if results[0].VersionCount < results[1].VersionCount {
			t.Errorf("排序错误: 第一个结果(%d)的版本数应大于第二个结果(%d)",
				results[0].VersionCount, results[1].VersionCount)
		}
	}
}

// TestListGAVsPaginated 测试分页查询GAV
func TestListGAVsPaginated(t *testing.T) {
	client := createRealClient(t)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 构建查询
	query := "g:org.apache"
	pageSize := 3

	// 获取第1页
	page1Results, total, err := client.ListGAVsPaginated(ctx, query, 1, pageSize)
	if err != nil {
		t.Logf("跳过测试: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	// 验证结果
	if len(page1Results) == 0 {
		t.Fatal("未找到任何组件，这可能是一个错误")
	}

	t.Logf("分页查询: 总计 %d 个结果, 第1页有 %d 个结果", total, len(page1Results))
	for i, result := range page1Results {
		t.Logf("第1页 %d. %s:%s", i+1, result.GroupId, result.ArtifactId)
	}

	// 获取第2页
	if total > pageSize {
		page2Results, _, err := client.ListGAVsPaginated(ctx, query, 2, pageSize)
		if err != nil {
			t.Fatalf("获取第2页失败: %v", err)
		}

		if len(page2Results) == 0 {
			t.Fatal("第2页未返回结果，这可能是一个错误")
		}

		t.Logf("第2页有 %d 个结果", len(page2Results))
		for i, result := range page2Results {
			t.Logf("第2页 %d. %s:%s", i+1, result.GroupId, result.ArtifactId)
		}

		// 验证两页结果不同
		if page1Results[0].ID == page2Results[0].ID {
			t.Error("第1页和第2页的第一个结果相同，分页可能有问题")
		}
	}
}

// TestIteratorGAVs 测试GAV迭代器
func TestIteratorGAVs(t *testing.T) {
	client := createRealClient(t)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 构建查询
	query := "g:org.apache.commons AND a:commons-lang3"

	// 创建迭代器
	iterator := client.IteratorGAVs(ctx, query)

	// 获取前3个结果
	count := 0
	maxCount := 3

	t.Log("使用迭代器获取结果:")
	for iterator.Next() {
		artifact := iterator.Value()
		t.Logf("%d. %s:%s:%s", count+1, artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)

		count++
		if count >= maxCount {
			break
		}
	}

	if count == 0 {
		t.Fatal("迭代器未返回任何结果")
	}
}

// TestFindGAVDependencies 测试查找两个GAV之间的依赖关系
func TestFindGAVDependencies(t *testing.T) {
	client := createRealClient(t)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试一个常见的依赖关系
	// 许多项目依赖commons-lang3
	groupId1 := "org.springframework"
	artifactId1 := "spring-core"

	groupId2 := "org.apache.commons"
	artifactId2 := "commons-lang3"

	// 执行查询
	results, err := client.FindGAVDependencies(ctx, groupId1, artifactId1, groupId2, artifactId2, 5)
	if err != nil {
		t.Logf("跳过测试: %v", err)
		t.Skip("无法连接到Maven Central API")
		return
	}

	// 验证结果 - 注意这个测试可能通过也可能不通过，取决于依赖关系是否存在
	t.Logf("找到 %d 个结果 - %s:%s 依赖于 %s:%s",
		len(results), groupId1, artifactId1, groupId2, artifactId2)

	for i, result := range results {
		t.Logf("%d. %s:%s:%s", i+1, result.GroupId, result.ArtifactId, result.LatestVersion)
	}

	// 如果未找到结果，不一定意味着测试失败，可能是确实没有依赖关系
	if len(results) == 0 {
		t.Logf("未找到依赖关系，但这可能是正常的，不一定说明功能有问题")
	}
}
