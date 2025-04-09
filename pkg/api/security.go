package api

import (
	"context"

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
func (c *Client) GetSecurityRating(ctx context.Context, groupId, artifactId, version string) (*response.SecurityRating, error) {
	// 构建请求
	securityRequest := request.NewSearchRequest().
		SetQuery(request.NewQuery().
			SetGroupId(groupId).
			SetArtifactId(artifactId).
			SetVersion(version)).
		AddCustomParam("fl", "id,g,a,v,p,ec,timestamp,tags,vulnerabilities").
		SetLimit(1)

	// 执行请求
	result, err := SearchRequestJsonDoc[map[string]interface{}](c, ctx, securityRequest)
	if err != nil {
		return nil, err
	}

	// 检查结果
	if len(result.ResponseBody.Docs) == 0 {
		return nil, ErrNotFound
	}

	// 解析安全信息
	doc := result.ResponseBody.Docs[0]
	rating := &response.SecurityRating{}

	// 解析漏洞信息
	if vulns, ok := doc["vulnerabilities"]; ok {
		if vulnList, ok := vulns.([]interface{}); ok {
			rating.VulnCount = len(vulnList)

			// 设置最高严重性级别
			if rating.VulnCount > 0 {
				rating.Severity = string(SecuritySeverityLow)
				for _, vuln := range vulnList {
					if vulnMap, ok := vuln.(map[string]interface{}); ok {
						if sev, ok := vulnMap["severity"].(string); ok {
							if isSeverityHigher(rating.Severity, sev) {
								rating.Severity = sev
							}
						}

						// 收集漏洞建议
						if advisory, ok := vulnMap["advisory"].(string); ok {
							rating.Advisories = append(rating.Advisories, advisory)
						}
					}
				}
			} else {
				rating.Severity = string(SecuritySeverityNone)
			}
		}
	}

	// 计算评分 (0-10分，越高越安全)
	switch SecuritySeverity(rating.Severity) {
	case SecuritySeverityCritical:
		rating.Score = 0.0
	case SecuritySeverityHigh:
		rating.Score = 2.5
	case SecuritySeverityMedium:
		rating.Score = 5.0
	case SecuritySeverityLow:
		rating.Score = 7.5
	case SecuritySeverityNone, "":
		rating.Score = 10.0
	}

	return rating, nil
}

// isSeverityHigher 判断给定的严重性是否高于当前严重性
func isSeverityHigher(currentSeverity, newSeverity string) bool {
	severityRank := map[string]int{
		string(SecuritySeverityCritical): 4,
		string(SecuritySeverityHigh):     3,
		string(SecuritySeverityMedium):   2,
		string(SecuritySeverityLow):      1,
		string(SecuritySeverityNone):     0,
		"":                               -1,
	}

	currentRank := severityRank[currentSeverity]
	newRank := severityRank[newSeverity]

	return newRank > currentRank
}

// SearchVulnerableArtifacts 搜索具有指定严重性级别或更高的漏洞的制品
func (c *Client) SearchVulnerableArtifacts(ctx context.Context, minSeverity SecuritySeverity, limit int) ([]*response.Artifact, error) {
	// 构建请求
	vulnQuery := request.NewQuery().
		SetCustomQuery("vulnerabilities.severity:" + string(minSeverity))

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
