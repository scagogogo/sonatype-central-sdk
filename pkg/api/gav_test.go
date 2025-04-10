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
