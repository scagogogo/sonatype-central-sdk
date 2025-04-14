package api

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// GetVersionInfo 获取组件的版本信息
//
// 该方法提供获取指定Maven制品特定版本详细信息的功能。返回的信息包括版本元数据、
// 发布时间、依赖项、许可证信息等。此方法适用于需要深入了解特定版本信息的场景，
// 如检查版本兼容性、获取版本发布日期或确认特定版本是否存在。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 制品的组ID，如"org.springframework"
//   - artifactId: 制品的制品ID，如"spring-core"
//   - version: 要查询的特定版本，如"5.3.25"
//
// 返回:
//   - *response.VersionInfo: 包含版本详细信息的对象
//   - error: 获取版本信息过程中发生的错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 获取Spring Core 5.3.25版本的详细信息
//	versionInfo, err := client.GetVersionInfo(ctx, "org.springframework", "spring-core", "5.3.25")
//	if err != nil {
//	    log.Fatalf("获取版本信息失败: %v", err)
//	}
//
//	// 输出版本信息
//	fmt.Printf("版本: %s:%s:%s\n", versionInfo.GroupId, versionInfo.ArtifactId, versionInfo.Version)
//	fmt.Printf("发布时间: %s\n", versionInfo.LastUpdated)
//	fmt.Printf("文件大小: %d 字节\n", versionInfo.FileSize)
//	fmt.Printf("打包类型: %s\n", versionInfo.Packaging)
//
//	// 检查POM文件
//	if versionInfo.HasPom {
//	    fmt.Println("提供POM文件")
//	}
//
//	// 输出可用文件列表
//	for _, file := range versionInfo.AvailableFiles {
//	    fmt.Printf("- %s (%s)\n", file.Name, file.Type)
//	}
func (c *Client) GetVersionInfo(ctx context.Context, groupId, artifactId, version string) (*response.VersionInfo, error) {
	// 构建URL
	targetUrl := fmt.Sprintf("%s/solrsearch/select?q=g:%s+AND+a:%s+AND+v:%s&rows=1&wt=json",
		c.baseURL,
		url.QueryEscape(groupId),
		url.QueryEscape(artifactId),
		url.QueryEscape(version))

	// 创建响应对象
	var result response.VersionInfo

	// 执行请求
	_, err := c.doRequest(ctx, "GET", targetUrl, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ListVersions 根据GroupID和artifactId列出下面的所有版本
//
// 该方法用于获取指定Maven制品的所有可用版本列表。结果按时间戳排序，最新版本排在前面。
// 这对于版本跟踪、兼容性检查、升级规划和历史版本分析非常有用。同时支持限制结果数量
// 或获取完整版本历史。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 制品的组ID，如"com.google.guava"
//   - artifactId: 制品的制品ID，如"guava"
//   - limit: 最大返回结果数量，如果小于等于0则返回所有版本
//
// 返回:
//   - []*response.Version: 制品版本列表，按发布时间降序排序
//   - error: 获取版本列表过程中发生的错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 获取Guava库的所有版本
//	versions, err := client.ListVersions(ctx, "com.google.guava", "guava", 0)
//	if err != nil {
//	    log.Fatalf("获取版本列表失败: %v", err)
//	}
//
//	// 输出版本信息
//	fmt.Printf("Guava库共有%d个版本:\n", len(versions))
//	for i, version := range versions {
//	    // 转换时间戳为可读格式
//	    releaseDate := time.Unix(version.Timestamp/1000, 0).Format("2006-01-02")
//	    fmt.Printf("%d. %s (发布于: %s)\n", i+1, version.Version, releaseDate)
//	}
//
//	// 仅获取最新的5个版本
//	recentVersions, err := client.ListVersions(ctx, "com.google.guava", "guava", 5)
//	if err != nil {
//	    log.Fatalf("获取最新版本失败: %v", err)
//	}
//
//	fmt.Println("\nGuava库最新的5个版本:")
//	for i, version := range recentVersions {
//	    fmt.Printf("%d. %s\n", i+1, version.Version)
//	}
func (c *Client) ListVersions(ctx context.Context, groupId, artifactId string, limit int) ([]*response.Version, error) {
	if limit <= 0 {
		return c.IteratorVersions(ctx, groupId, artifactId).ToSlice()
	} else {
		search := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId(groupId).SetArtifactId(artifactId)).SetCore("gav").SetLimit(limit)
		result, err := SearchRequestJsonDoc[*response.Version](c, ctx, search)
		if err != nil {
			return nil, err
		}
		if result == nil || result.ResponseBody == nil {
			return nil, errors.New("empty response body")
		}
		return result.ResponseBody.Docs, nil
	}
}

// IteratorVersions 返回一个版本迭代器
//
// 该方法提供一个迭代器接口，用于高效处理指定Maven制品的所有版本列表。
// 使用迭代器模式可以分页获取大量版本数据，避免一次性加载全部版本到内存中，
// 特别适合版本数量非常多的制品，或需要按需处理版本信息的场景。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 制品的组ID，如"org.apache.logging.log4j"
//   - artifactId: 制品的制品ID，如"log4j-core"
//
// 返回:
//   - *SearchIterator[*response.Version]: 用于遍历版本列表的迭代器对象
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 创建版本迭代器
//	iterator := client.IteratorVersions(ctx, "org.apache.logging.log4j", "log4j-core")
//
//	// 配置迭代器
//	iterator.SetPageSize(50) // 每页50个结果
//
//	// 使用迭代器处理所有版本
//	versionCount := 0
//	releaseYears := make(map[int]int) // 统计每年发布的版本数量
//
//	for iterator.HasNext() {
//	    versions, err := iterator.Next()
//	    if err != nil {
//	        log.Printf("获取下一批版本失败: %v", err)
//	        continue
//	    }
//
//	    for _, version := range versions {
//	        versionCount++
//	        // 分析版本发布年份
//	        year := time.Unix(version.Timestamp/1000, 0).Year()
//	        releaseYears[year]++
//	    }
//	}
//
//	// 输出统计结果
//	fmt.Printf("总版本数: %d\n", versionCount)
//	fmt.Println("按年份统计版本数量:")
//	for year, count := range releaseYears {
//	    fmt.Printf("%d年: %d个版本\n", year, count)
//	}
func (c *Client) IteratorVersions(ctx context.Context, groupId, artifactId string) *SearchIterator[*response.Version] {
	search := request.NewSearchRequest().SetQuery(request.NewQuery().SetGroupId(groupId).SetArtifactId(artifactId)).SetCore("gav")
	return NewSearchIterator[*response.Version](search).WithClient(c)
}

// GetLatestVersion 获取最新的发布版本
//
// 该方法提供了快速获取指定Maven制品最新发布版本的便捷方式。它返回按时间戳排序的
// 第一个版本，通常是制品的最新稳定版本。这对于确保使用最新版本、检查更新或获取
// 当前推荐版本非常有用。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - groupId: 制品的组ID，如"org.springframework"
//   - artifactId: 制品的制品ID，如"spring-core"
//
// 返回:
//   - *response.Version: 最新版本的详细信息对象
//   - error: 获取最新版本过程中发生的错误，如果制品不存在也会返回错误
//
// 使用示例:
//
//	client := api.NewClient()
//	ctx := context.Background()
//
//	// 获取Spring Core最新版本
//	latestVersion, err := client.GetLatestVersion(ctx, "org.springframework", "spring-core")
//	if err != nil {
//	    log.Fatalf("获取最新版本失败: %v", err)
//	}
//
//	// 输出版本信息
//	fmt.Printf("Spring Core最新版本: %s\n", latestVersion.Version)
//	releaseDate := time.Unix(latestVersion.Timestamp/1000, 0).Format("2006-01-02")
//	fmt.Printf("发布日期: %s\n", releaseDate)
//	fmt.Printf("文件大小: %d字节\n", latestVersion.FileSize)
//
//	// 使用最新版本信息
//	downloadUrl := fmt.Sprintf(
//	    "https://repo1.maven.org/maven2/%s/%s/%s/%s-%s.jar",
//	    strings.ReplaceAll(latestVersion.GroupId, ".", "/"),
//	    latestVersion.ArtifactId,
//	    latestVersion.Version,
//	    latestVersion.ArtifactId,
//	    latestVersion.Version)
//	fmt.Printf("下载地址: %s\n", downloadUrl)
func (c *Client) GetLatestVersion(ctx context.Context, groupId, artifactId string) (*response.Version, error) {
	versions, err := c.ListVersions(ctx, groupId, artifactId, 1)
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for %s:%s", groupId, artifactId)
	}
	return versions[0], nil
}

// GetVersionsWithMetadata 获取所有版本并附带元数据信息
func (c *Client) GetVersionsWithMetadata(ctx context.Context, groupId, artifactId string) ([]*response.VersionWithMetadata, error) {
	versions, err := c.ListVersions(ctx, groupId, artifactId, 0)
	if err != nil {
		return nil, err
	}

	result := make([]*response.VersionWithMetadata, 0, len(versions))
	for _, version := range versions {
		versionInfo, err := c.GetVersionInfo(ctx, groupId, artifactId, version.Version)
		if err != nil {
			continue
		}

		result = append(result, &response.VersionWithMetadata{
			Version:     version,
			VersionInfo: versionInfo,
		})
	}

	return result, nil
}

// FilterVersions 根据条件过滤版本
func (c *Client) FilterVersions(ctx context.Context, groupId, artifactId string, filter func(*response.Version) bool) ([]*response.Version, error) {
	versions, err := c.ListVersions(ctx, groupId, artifactId, 0)
	if err != nil {
		return nil, err
	}

	result := make([]*response.Version, 0)
	for _, version := range versions {
		if filter(version) {
			result = append(result, version)
		}
	}

	return result, nil
}

// CompareVersions 比较两个版本
func (c *Client) CompareVersions(ctx context.Context, groupId, artifactId string, version1, version2 string) (*response.VersionComparison, error) {
	v1Info, err := c.GetVersionInfo(ctx, groupId, artifactId, version1)
	if err != nil {
		return nil, err
	}

	v2Info, err := c.GetVersionInfo(ctx, groupId, artifactId, version2)
	if err != nil {
		return nil, err
	}

	return &response.VersionComparison{
		Version1:    version1,
		Version2:    version2,
		V1Timestamp: v1Info.LastUpdated,
		V2Timestamp: v2Info.LastUpdated,
	}, nil
}

// HasVersion 检查特定版本是否存在
func (c *Client) HasVersion(ctx context.Context, groupId, artifactId, version string) (bool, error) {
	_, err := c.GetVersionInfo(ctx, groupId, artifactId, version)
	if err != nil {
		// 使用errors.Is检查是否是NotFound错误
		if errors.Is(err, ErrNotFound) {
			return false, nil
		}
		// 其他错误返回给调用者
		return false, err
	}
	return true, nil
}
