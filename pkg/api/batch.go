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
