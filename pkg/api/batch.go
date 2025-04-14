package api

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// BatchDownloadResult 批量下载结果
type BatchDownloadResult struct {
	// 文件路径
	FilePath string

	// 本地保存路径
	LocalPath string

	// 是否成功
	Success bool

	// 错误信息
	Error error

	// 文件大小
	Size int
}

// BatchDownloadFiles 批量下载文件到本地目录
//
// 此方法允许同时下载多个文件到指定的本地路径。它使用goroutine并行处理下载任务，
// 提高大批量文件下载时的效率。每个下载任务的结果将包含成功/失败状态和相关信息。
//
// 参数:
//   - ctx: 请求的上下文，用于控制请求的生命周期
//   - fileMappings: 远程文件路径到本地保存路径的映射，key为远程路径，value为本地路径
//
// 返回值:
//   - []BatchDownloadResult: 每个文件的下载结果，包含成功/失败状态、错误信息和文件大小等信息
//
// 示例:
//
//	client := sonatype.NewClient()
//	mappings := map[string]string{
//		"org/example/lib/1.0.0/lib-1.0.0.jar": "/downloads/lib-1.0.0.jar",
//		"org/example/app/2.0.0/app-2.0.0.jar": "/downloads/app-2.0.0.jar",
//	}
//	results := client.BatchDownloadFiles(context.Background(), mappings)
//	for _, result := range results {
//		if result.Success {
//			fmt.Printf("下载成功: %s (大小: %d 字节)\n", result.LocalPath, result.Size)
//		} else {
//			fmt.Printf("下载失败: %s (错误: %v)\n", result.FilePath, result.Error)
//		}
//	}
func (c *Client) BatchDownloadFiles(ctx context.Context, fileMappings map[string]string) []BatchDownloadResult {
	var (
		wg      sync.WaitGroup
		results = make([]BatchDownloadResult, 0, len(fileMappings))
		mu      sync.Mutex
	)

	// 为每个文件创建下载任务
	for remotePath, localPath := range fileMappings {
		wg.Add(1)

		go func(remotePath, localPath string) {
			defer wg.Done()

			result := BatchDownloadResult{
				FilePath:  remotePath,
				LocalPath: localPath,
			}

			// 下载文件
			data, err := c.Download(ctx, remotePath)
			if err != nil {
				result.Success = false
				result.Error = err

				mu.Lock()
				results = append(results, result)
				mu.Unlock()
				return
			}

			// 确保目录存在
			dir := filepath.Dir(localPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				result.Success = false
				result.Error = err

				mu.Lock()
				results = append(results, result)
				mu.Unlock()
				return
			}

			// 保存文件
			if err := os.WriteFile(localPath, data, 0644); err != nil {
				result.Success = false
				result.Error = err

				mu.Lock()
				results = append(results, result)
				mu.Unlock()
				return
			}

			result.Success = true
			result.Size = len(data)

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(remotePath, localPath)
	}

	// 等待所有下载完成
	wg.Wait()
	return results
}

// BatchSearchArtifacts 批量搜索制品
//
// 此方法允许同时执行多个搜索请求，基于提供的搜索条件和搜索类型。
// 它使用goroutine并行处理搜索任务，提高批量搜索时的效率。
//
// 参数:
//   - ctx: 请求的上下文，用于控制请求的生命周期
//   - searchCriteria: 搜索条件列表，每个条件将作为一个单独的搜索任务
//   - searchType: 搜索类型，支持"groupId"、"artifactId"、"className"、"tag"等
//   - limit: 每个搜索返回的最大结果数
//
// 返回值:
//   - map[string][]*response.Artifact: 搜索结果映射，key为搜索条件，value为匹配的制品列表
//
// 示例:
//
//	client := sonatype.NewClient()
//	criteria := []string{"org.apache.commons", "org.springframework", "com.google.guava"}
//	results := client.BatchSearchArtifacts(context.Background(), criteria, "groupId", 10)
//	for criteria, artifacts := range results {
//		fmt.Printf("搜索条件 %s 找到 %d 个制品:\n", criteria, len(artifacts))
//		for _, artifact := range artifacts {
//			fmt.Printf("  - %s:%s:%s\n", artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
//		}
//	}
func (c *Client) BatchSearchArtifacts(ctx context.Context, searchCriteria []string, searchType string, limit int) map[string][]*response.Artifact {
	var (
		wg      sync.WaitGroup
		results = make(map[string][]*response.Artifact)
		mu      sync.Mutex
	)

	// 为每个搜索条件创建任务
	for _, criteria := range searchCriteria {
		wg.Add(1)

		go func(criteria string) {
			defer wg.Done()

			var artifacts []*response.Artifact
			var err error

			// 根据搜索类型执行不同的搜索
			switch searchType {
			case "groupId":
				artifacts, err = c.SearchByGroupId(ctx, criteria, limit)
			case "artifactId":
				artifacts, err = c.SearchByArtifactId(ctx, criteria, limit)
			case "className":
				// 这里需要转换返回类型
				versions, err := c.SearchByClassName(ctx, criteria, limit)
				if err == nil && len(versions) > 0 {
					// 将版本信息转换为制品信息（简化处理）
					for _, v := range versions {
						artifacts = append(artifacts, &response.Artifact{
							ID:            v.ID,
							GroupId:       v.GroupId,
							ArtifactId:    v.ArtifactId,
							LatestVersion: v.Version,
							Packaging:     v.Packaging,
							Timestamp:     v.Timestamp,
						})
					}
				}
			case "tag":
				artifacts, err = c.SearchByTag(ctx, criteria, limit)
			}

			if err == nil {
				mu.Lock()
				results[criteria] = artifacts
				mu.Unlock()
			}
		}(criteria)
	}

	// 等待所有搜索完成
	wg.Wait()
	return results
}

// BatchDownloadDependencies 批量下载制品的依赖项
//
// 此方法用于下载指定制品的所有依赖项到本地目录。它会先获取制品的元数据信息，
// 然后解析出所有依赖项并下载。对于非可选依赖，会同时下载JAR文件和POM文件。
//
// 参数:
//   - ctx: 请求的上下文，用于控制请求的生命周期
//   - groupId: 制品的组ID
//   - artifactId: 制品的ID
//   - version: 制品的版本
//   - outputDir: 依赖项文件的输出目录
//
// 返回值:
//   - []BatchDownloadResult: 依赖项的下载结果，包含成功/失败状态、错误信息和文件大小等信息
//   - error: 如果获取依赖元数据时发生错误，则返回此错误
//
// 示例:
//
//	client := sonatype.NewClient()
//	results, err := client.BatchDownloadDependencies(
//		context.Background(),
//		"org.apache.commons",
//		"commons-lang3",
//		"3.12.0",
//		"/dependencies"
//	)
//	if err != nil {
//		fmt.Printf("获取依赖失败: %v\n", err)
//		return
//	}
//	fmt.Printf("共下载了 %d 个依赖项\n", len(results))
//	for _, result := range results {
//		if result.Success {
//			fmt.Printf("下载成功: %s\n", result.LocalPath)
//		} else {
//			fmt.Printf("下载失败: %s (错误: %v)\n", result.FilePath, result.Error)
//		}
//	}
func (c *Client) BatchDownloadDependencies(ctx context.Context, groupId, artifactId, version, outputDir string) ([]BatchDownloadResult, error) {
	// 获取制品元数据
	metadata, err := c.GetArtifactMetadata(ctx, groupId, artifactId, version)
	if err != nil {
		return nil, err
	}

	// 如果没有依赖，直接返回
	if len(metadata.Dependencies) == 0 {
		return []BatchDownloadResult{}, nil
	}

	// 准备下载映射
	fileMappings := make(map[string]string)

	for _, dep := range metadata.Dependencies {
		if dep.GroupId != "" && dep.ArtifactId != "" && dep.Version != "" {
			// 构建POM路径
			remotePath := BuildArtifactPath(dep.GroupId, dep.ArtifactId, dep.Version, POM)

			// 构建本地保存路径
			localDir := filepath.Join(outputDir, dep.GroupId, dep.ArtifactId, dep.Version)
			localPath := filepath.Join(localDir, dep.ArtifactId+"-"+dep.Version+".pom")

			fileMappings[remotePath] = localPath

			// 如果不是可选依赖，也下载JAR
			if !dep.Optional {
				remoteJarPath := BuildArtifactPath(dep.GroupId, dep.ArtifactId, dep.Version, JAR)
				localJarPath := filepath.Join(localDir, dep.ArtifactId+"-"+dep.Version+".jar")

				fileMappings[remoteJarPath] = localJarPath
			}
		}
	}

	// 执行批量下载
	return c.BatchDownloadFiles(ctx, fileMappings), nil
}
