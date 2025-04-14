package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SecuritySeverity 安全严重性级别
type SecuritySeverity string

const (
	SecuritySeverityCritical SecuritySeverity = "CRITICAL"
	SecuritySeverityHigh     SecuritySeverity = "HIGH"
	SecuritySeverityMedium   SecuritySeverity = "MEDIUM"
	SecuritySeverityLow      SecuritySeverity = "LOW"
	SecuritySeverityNone     SecuritySeverity = "NONE"
)

// GetSecurityRating 获取制品的安全评分
//
// 此方法提供了一种获取指定制品版本安全评分的方式，返回包含安全评级和相关详情的安全评分信息。
// 安全评分基于该制品版本已知的漏洞数量和严重性来计算，可用于评估使用该制品的潜在风险。
//
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - groupId: Maven坐标中的组ID
//   - artifactId: Maven坐标中的制品ID
//   - version: 制品的版本号
//
// 返回:
//   - *response.SecurityRating: 包含安全评级、评分和相关详情的安全评分信息
//   - error: 如果获取安全评分失败，返回错误信息
//
// 例子:
//
//	// 初始化客户端
//	client := sonatype.NewClient("https://repo1.maven.org/maven2")
//
//	// 获取指定制品版本的安全评分
//	rating, err := client.GetSecurityRating(ctx, "org.apache.commons", "commons-lang3", "3.12.0")
//	if err != nil {
//	    log.Fatalf("获取安全评分失败: %v", err)
//	}
//
//	// 使用安全评分信息
//	fmt.Printf("安全评级: %s\n", rating.Level)
//	fmt.Printf("安全评分: %.2f\n", rating.Score)
//	if rating.Score > 7.0 {
//	    fmt.Println("警告: 该制品存在高风险漏洞，建议升级或更换替代品")
//	}
func (c *Client) GetSecurityRating(ctx context.Context, groupId, artifactId, version string) (*response.SecurityRating, error) {
	targetUrl := fmt.Sprintf("%s/api/security/rating/%s/%s/%s", c.baseURL, groupId, artifactId, version)
	var securityRating response.SecurityRating
	_, err := c.doRequest(ctx, "GET", targetUrl, nil, &securityRating)
	if err != nil {
		return nil, err
	}
	return &securityRating, nil
}

// SearchVulnerableArtifacts 搜索Maven中央仓库中具有已知漏洞的构件。
//
// 该方法允许使用自定义查询来搜索具有安全漏洞的构件。通常，查询中应包含
// 漏洞相关的条件，例如 "vulnerabilities.severity:[MODERATE TO CRITICAL]"
// 来过滤特定严重性级别的漏洞。
//
// 参数:
//   - ctx: 请求的上下文，用于控制请求的生命周期。
//   - query: 搜索查询参数，可以指定过滤条件如漏洞级别、构件名称等。
//
// 返回:
//   - *response.Response[response.ArtifactMetadata]: 包含搜索结果的响应对象。
//   - error: 如果搜索失败，返回相应的错误。
//
// 示例:
//
//	client := api.NewClient(api.ClientOptions{})
//
//	// 创建一个查询，搜索具有严重或关键级别漏洞的Spring相关构件
//	query := request.NewQuery().
//		SetCustomQuery("vulnerabilities.severity:[HIGH TO CRITICAL]").
//		SetArtifactId("spring")
//
//	// 执行搜索
//	results, err := client.SearchVulnerableArtifacts(context.Background(), query)
//	if err != nil {
//		log.Fatalf("搜索失败: %v", err)
//	}
//
//	// 处理结果
//	fmt.Printf("找到 %d 个具有漏洞的构件\n", results.Response.NumFound)
//	for _, artifact := range results.Response.Docs {
//		fmt.Printf("构件 %s:%s 含有漏洞\n", artifact.GroupId, artifact.ArtifactId)
//		// 可以进一步检查 artifact.Vulnerabilities 字段获取漏洞详情
//	}
func (c *Client) SearchVulnerableArtifacts(ctx context.Context, query *request.Query) (*response.Response[response.ArtifactMetadata], error) {
	// 构建请求
	// 使用传入的查询参数
	searchRequest := request.NewSearchRequest().
		SetQuery(query).
		AddCustomParam("fl", "id,g,a,latestVersion,p,timestamp,versionCount,text,ec,vulnerabilities").
		SetSort("vulnerabilities.severity", false) // 按严重性降序排序

	// 执行请求
	result, err := SearchRequestJsonDoc[response.ArtifactMetadata](c, ctx, searchRequest)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetVulnerabilityDetails 获取特定构件版本的漏洞详情
func (c *Client) GetVulnerabilityDetails(ctx context.Context, groupId, artifactId, version string) (*response.VulnerabilityDetails, error) {
	targetUrl := fmt.Sprintf("%s/api/security/vulnerabilities/%s/%s/%s", c.baseURL, groupId, artifactId, version)
	var details response.VulnerabilityDetails
	_, err := c.doRequest(ctx, "GET", targetUrl, nil, &details)
	if err != nil {
		return nil, err
	}
	return &details, nil
}

// CheckCVEImpact 检查特定构件是否受到某个CVE编号漏洞的影响
func (c *Client) CheckCVEImpact(ctx context.Context, cveId, groupId, artifactId, version string) (bool, *response.Vulnerability, error) {
	details, err := c.GetVulnerabilityDetails(ctx, groupId, artifactId, version)
	if err != nil {
		return false, nil, err
	}

	for _, vuln := range details.Vulnerabilities {
		if strings.EqualFold(vuln.CVE, cveId) {
			return true, vuln, nil
		}
	}

	return false, nil, nil
}

// FindArtifactsByCVE 根据CVE编号查找受影响的构件
func (c *Client) FindArtifactsByCVE(ctx context.Context, cveId string, limit int) ([]*response.Artifact, error) {
	// 构建请求
	vulnQuery := request.NewQuery().
		SetCustomQuery(fmt.Sprintf("cve:%s", cveId))

	vulnRequest := request.NewSearchRequest().
		SetQuery(vulnQuery).
		AddCustomParam("fl", "id,g,a,latestVersion,p,timestamp,versionCount,text,ec,vulnerabilities").
		SetSort("vulnerabilities.severity", false). // 按严重性降序排序
		SetLimit(limit)

	// 执行请求
	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, vulnRequest)
	if err != nil {
		return nil, err
	}

	return result.ResponseBody.Docs, nil
}

// CompareVersionSecurity 比较两个版本的安全性差异
func (c *Client) CompareVersionSecurity(ctx context.Context, groupId, artifactId, version1, version2 string) (*response.SecurityComparison, error) {
	rating1, err := c.GetSecurityRating(ctx, groupId, artifactId, version1)
	if err != nil {
		return nil, fmt.Errorf("获取版本1安全评分失败: %v", err)
	}

	rating2, err := c.GetSecurityRating(ctx, groupId, artifactId, version2)
	if err != nil {
		return nil, fmt.Errorf("获取版本2安全评分失败: %v", err)
	}

	comparison := &response.SecurityComparison{
		GroupId:         groupId,
		ArtifactId:      artifactId,
		Version1:        version1,
		Version2:        version2,
		Rating1:         rating1,
		Rating2:         rating2,
		SaferVersion:    version1,
		ScoreDifference: 0,
	}

	// 计算分数差异（较高分数表示较高风险）
	if rating1.Score > rating2.Score {
		comparison.SaferVersion = version2
		comparison.ScoreDifference = rating1.Score - rating2.Score
	} else if rating2.Score > rating1.Score {
		comparison.SaferVersion = version1
		comparison.ScoreDifference = rating2.Score - rating1.Score
	}

	return comparison, nil
}

// GetRecommendedSecureVersion 获取修复特定漏洞的推荐版本
func (c *Client) GetRecommendedSecureVersion(ctx context.Context, groupId, artifactId, currentVersion string) (string, error) {
	// 获取当前版本的漏洞信息
	vulnDetails, err := c.GetVulnerabilityDetails(ctx, groupId, artifactId, currentVersion)
	if err != nil {
		return "", err
	}

	// 如果没有漏洞，则当前版本已经是安全的
	if len(vulnDetails.Vulnerabilities) == 0 {
		return currentVersion, nil
	}

	// 获取所有版本
	versions, err := c.ListVersions(ctx, groupId, artifactId, 100)
	if err != nil {
		return "", err
	}

	// 找到比当前版本更新的版本
	var newerVersions []string
	foundCurrent := false
	for i := len(versions) - 1; i >= 0; i-- {
		v := versions[i].Version
		if v == currentVersion {
			foundCurrent = true
			continue
		}
		if foundCurrent {
			newerVersions = append(newerVersions, v)
		}
	}

	// 如果没有更新的版本，则返回当前版本
	if len(newerVersions) == 0 {
		return currentVersion, nil
	}

	// 检查每个更新的版本，找到第一个没有漏洞的版本
	for _, v := range newerVersions {
		details, err := c.GetVulnerabilityDetails(ctx, groupId, artifactId, v)
		if err != nil {
			continue
		}
		if len(details.Vulnerabilities) == 0 {
			return v, nil
		}
	}

	// 如果所有更新的版本都有漏洞，则返回当前版本
	return currentVersion, nil
}

// BatchSecurityScan 批量检查多个构件的安全状态
func (c *Client) BatchSecurityScan(ctx context.Context, artifacts []*response.ArtifactRef) ([]*response.SecurityScanResult, error) {
	var results []*response.SecurityScanResult

	for _, artifact := range artifacts {
		result := &response.SecurityScanResult{
			GroupId:    artifact.GroupId,
			ArtifactId: artifact.ArtifactId,
			Version:    artifact.Version,
		}

		rating, err := c.GetSecurityRating(ctx, artifact.GroupId, artifact.ArtifactId, artifact.Version)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.SecurityRating = rating
		}

		results = append(results, result)
	}

	return results, nil
}

// GetVulnerabilityTimeline 获取构件漏洞随版本变化的时间线
func (c *Client) GetVulnerabilityTimeline(ctx context.Context, groupId, artifactId string, maxVersions int) (*response.VulnerabilityTimeline, error) {
	// 获取所有版本
	versions, err := c.ListVersions(ctx, groupId, artifactId, maxVersions)
	if err != nil {
		return nil, err
	}

	timeline := &response.VulnerabilityTimeline{
		GroupId:    groupId,
		ArtifactId: artifactId,
		Entries:    make([]*response.TimelineEntry, 0, len(versions)),
	}

	var previousScore float64
	var previousEntry *response.TimelineEntry

	// 从最旧的版本开始分析
	for i := len(versions) - 1; i >= 0; i-- {
		version := versions[i]

		// 获取版本的安全评分
		rating, err := c.GetSecurityRating(ctx, groupId, artifactId, version.Version)
		if err != nil {
			continue
		}

		// 获取版本的漏洞详情
		vulnDetails, err := c.GetVulnerabilityDetails(ctx, groupId, artifactId, version.Version)
		if err != nil {
			continue
		}

		entry := &response.TimelineEntry{
			Version:   version.Version,
			Timestamp: version.Timestamp, // 使用版本的时间戳
			VulnCount: len(vulnDetails.Vulnerabilities),
			Severity:  rating.Severity,
			Score:     rating.Score,
			Change:    "STABLE",
		}

		// 与前一个版本比较
		if previousEntry != nil {
			if entry.Score < previousScore {
				entry.Change = "IMPROVED"
				entry.ChangeDetails = fmt.Sprintf("安全评分从 %.2f 提升到 %.2f", previousScore, entry.Score)
			} else if entry.Score > previousScore {
				entry.Change = "DEGRADED"
				entry.ChangeDetails = fmt.Sprintf("安全评分从 %.2f 降低到 %.2f", previousScore, entry.Score)
			}
		}

		timeline.Entries = append(timeline.Entries, entry)
		previousScore = entry.Score
		previousEntry = entry
	}

	return timeline, nil
}

// GetComponentVulnerabilityOverview 获取组件的漏洞概览，包括不同版本的安全状态
func (c *Client) GetComponentVulnerabilityOverview(ctx context.Context, groupId, artifactId string, limitVersions int) (*response.ComponentVulnOverview, error) {
	// 获取组件的版本列表
	versions, err := c.ListVersions(ctx, groupId, artifactId, limitVersions)
	if err != nil {
		return nil, fmt.Errorf("获取版本列表失败: %w", err)
	}

	overview := &response.ComponentVulnOverview{
		GroupId:               groupId,
		ArtifactId:            artifactId,
		TotalVersions:         len(versions),
		VulnerableVersions:    0,
		LatestVersion:         "",
		LatestVulnFreeVersion: "",
		VersionRatings:        make(map[string]*response.SecurityRating),
		SeverityCounts:        make(map[string]int),
	}

	if len(versions) > 0 {
		overview.LatestVersion = versions[0].Version
	}

	// 评估每个版本
	for _, version := range versions {
		rating, err := c.GetSecurityRating(ctx, groupId, artifactId, version.Version)
		if err != nil {
			// 如果获取安全评分失败，记录错误并继续
			continue
		}

		overview.VersionRatings[version.Version] = rating

		// 更新漏洞版本计数
		if rating.VulnCount > 0 {
			overview.VulnerableVersions++

			// 更新不同严重性级别的计数
			if _, exists := overview.SeverityCounts[rating.Severity]; !exists {
				overview.SeverityCounts[rating.Severity] = 0
			}
			overview.SeverityCounts[rating.Severity]++
		} else if overview.LatestVulnFreeVersion == "" {
			// 记录最新的无漏洞版本
			overview.LatestVulnFreeVersion = version.Version
		}
	}

	return overview, nil
}

// FindSimilarVulnerableArtifacts 查找与指定组件有相似漏洞的其他组件
func (c *Client) FindSimilarVulnerableArtifacts(ctx context.Context, groupId, artifactId, version string, limit int) ([]*response.Artifact, error) {
	// 获取当前组件的漏洞信息
	vulnDetails, err := c.GetVulnerabilityDetails(ctx, groupId, artifactId, version)
	if err != nil {
		return nil, fmt.Errorf("获取组件漏洞信息失败: %w", err)
	}

	// 如果没有漏洞，则无法查找相似组件
	if len(vulnDetails.Vulnerabilities) == 0 {
		return nil, nil
	}

	// 收集CVE编号
	var cveList []string
	for _, vuln := range vulnDetails.Vulnerabilities {
		if vuln.CVE != "" {
			cveList = append(cveList, vuln.CVE)
		}
	}

	if len(cveList) == 0 {
		return nil, nil
	}

	// 构建查询，查找具有相同CVE的其他组件
	cveQuery := strings.Join(cveList, " OR ")
	query := request.NewQuery().
		SetCustomQuery("cve:(" + cveQuery + ")")

	// 排除当前组件
	excludeFilter := fmt.Sprintf("(g:%s AND a:%s)", groupId, artifactId)
	query.SetCustomQuery(query.ToRequestParamValue() + " AND -" + excludeFilter)

	searchRequest := request.NewSearchRequest().
		SetQuery(query).
		AddCustomParam("fl", "id,g,a,latestVersion,p,timestamp,versionCount,text,ec,vulnerabilities").
		SetSort("vulnerabilities.severity", false). // 按严重性降序排序
		SetLimit(limit)

	// 执行请求
	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, searchRequest)
	if err != nil {
		return nil, err
	}

	return result.ResponseBody.Docs, nil
}
