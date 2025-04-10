package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// LicenseType 定义了常见的开源许可证类型
type LicenseType string

const (
	LicenseTypeApache2   LicenseType = "Apache-2.0"
	LicenseTypeMIT       LicenseType = "MIT"
	LicenseTypeGPLv2     LicenseType = "GPL-2.0"
	LicenseTypeGPLv3     LicenseType = "GPL-3.0"
	LicenseTypeLGPLv2    LicenseType = "LGPL-2.0"
	LicenseTypeLGPLv3    LicenseType = "LGPL-3.0"
	LicenseTypeBSD2      LicenseType = "BSD-2-Clause"
	LicenseTypeBSD3      LicenseType = "BSD-3-Clause"
	LicenseTypeMPL       LicenseType = "MPL-2.0"
	LicenseTypeEPL       LicenseType = "EPL-2.0"
	LicenseTypeCDDL      LicenseType = "CDDL-1.0"
	LicenseTypeUnlicense LicenseType = "Unlicense"
)

// LicenseCategory 定义了许可证的类别
type LicenseCategory string

const (
	LicenseCategoryPermissive    LicenseCategory = "permissive"     // 宽松许可证，如MIT, Apache
	LicenseCategoryCopyleft      LicenseCategory = "copyleft"       // 传染性许可证，如GPL
	LicenseCategoryWeakCopyleft  LicenseCategory = "weak-copyleft"  // 弱传染性许可证，如LGPL
	LicenseCategoryNonCommercial LicenseCategory = "non-commercial" // 非商业许可证
)

// GetComponentLicenses 获取一个组件的许可证信息
func (c *Client) GetComponentLicenses(ctx context.Context, groupID, artifactID, version string) ([]response.LicenseInfo, error) {
	// 构建请求URL
	q := fmt.Sprintf("g:%s+AND+a:%s+AND+v:%s",
		url.QueryEscape(groupID), url.QueryEscape(artifactID), url.QueryEscape(version))

	// 创建查询
	query := request.NewQuery().SetCustomQuery(q)
	searchReq := request.NewSearchRequest().SetQuery(query)

	// 执行查询
	var resp response.Response[map[string]interface{}]
	err := c.SearchRequest(ctx, searchReq, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get license information: %w", err)
	}

	if resp.ResponseBody.NumFound == 0 {
		return nil, fmt.Errorf("component %s:%s:%s not found", groupID, artifactID, version)
	}

	// 解析文档中的许可证信息
	var licenses []response.LicenseInfo
	for _, doc := range resp.ResponseBody.Docs {
		if licField, ok := doc["licenseList"]; ok {
			if licList, ok := licField.([]interface{}); ok {
				for _, lic := range licList {
					licStr, ok := lic.(string)
					if !ok {
						continue
					}

					// 解析许可证信息
					licenses = append(licenses, parseLicense(licStr))
				}
			}
		}
	}

	return licenses, nil
}

// SearchByLicenseType 搜索使用特定许可证类型的组件
func (c *Client) SearchByLicenseType(ctx context.Context, licenseType LicenseType, limit int) ([]response.ArtifactRef, error) {
	// 构建查询请求
	q := fmt.Sprintf("l:%s", url.QueryEscape(string(licenseType)))
	query := request.NewQuery().SetCustomQuery(q)
	searchReq := request.NewSearchRequest().
		SetQuery(query).
		SetRows(limit)

	// 执行查询
	var resp response.Response[map[string]interface{}]
	err := c.SearchRequest(ctx, searchReq, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to search by license type: %w", err)
	}

	// 处理结果
	artifacts := make([]response.ArtifactRef, 0, len(resp.ResponseBody.Docs))
	for _, doc := range resp.ResponseBody.Docs {
		groupID, _ := doc["g"].(string)
		artifactID, _ := doc["a"].(string)
		version, _ := doc["v"].(string)

		artifacts = append(artifacts, response.ArtifactRef{
			GroupId:    groupID,
			ArtifactId: artifactID,
			Version:    version,
		})
	}

	return artifacts, nil
}

// FindLicenseConflicts 检查组件依赖项中的许可证冲突
func (c *Client) FindLicenseConflicts(ctx context.Context, artifacts []response.ArtifactRef) (*response.LicenseSummary, error) {
	if len(artifacts) == 0 {
		return &response.LicenseSummary{}, nil
	}

	// 保存所有发现的许可证
	foundLicenses := make(map[response.ArtifactRef][]response.LicenseInfo)
	licenseDistribution := make(map[string]int)
	categoryDistribution := make(map[string]int)
	artifactsByLicense := make(map[string][]response.ArtifactRef)

	// 获取每个组件的许可证信息
	for _, artifact := range artifacts {
		licenses, err := c.GetComponentLicenses(ctx, artifact.GroupId, artifact.ArtifactId, artifact.Version)
		if err != nil {
			continue // 跳过无法获取许可证信息的组件
		}

		foundLicenses[artifact] = licenses

		// 更新许可证分布统计
		for _, license := range licenses {
			licenseDistribution[license.Type]++
			categoryDistribution[license.Category]++

			// 更新按许可证分类的组件列表
			if artifactsByLicense[license.Type] == nil {
				artifactsByLicense[license.Type] = []response.ArtifactRef{}
			}
			artifactsByLicense[license.Type] = append(artifactsByLicense[license.Type], artifact)
		}
	}

	// 检查许可证冲突
	conflicts := findConflicts(foundLicenses)

	return &response.LicenseSummary{
		TotalArtifacts:       len(artifacts),
		LicenseDistribution:  licenseDistribution,
		CategoryDistribution: categoryDistribution,
		PotentialConflicts:   conflicts,
		ArtifactsByLicense:   artifactsByLicense,
	}, nil
}

// GetPopularLicenses 获取按使用频率排序的流行许可证
func (c *Client) GetPopularLicenses(ctx context.Context, limit int) (map[string]int, error) {
	// 使用facet查询获取许可证分布
	query := request.NewQuery().SetCustomQuery("*:*")
	searchReq := request.NewSearchRequest().
		SetQuery(query).
		AddCustomParam("facet", "true").
		AddCustomParam("facet.field", "l").
		AddCustomParam("facet.limit", fmt.Sprintf("%d", limit)).
		SetRows(0) // 只需要聚合结果，不需要文档

	// 执行查询
	var result response.Response[json.RawMessage]
	err := c.SearchRequest(ctx, searchReq, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular licenses: %w", err)
	}

	// 处理facet结果
	licenses := make(map[string]int)

	if result.FacetCounts != nil && result.FacetCounts.FacetFields != nil {
		if licenseField, ok := result.FacetCounts.FacetFields["l"]; ok {
			// facet结果格式为[license1, count1, license2, count2, ...]
			for i := 0; i < len(licenseField); i += 2 {
				if licName, ok := licenseField[i].(string); ok {
					if count, ok := licenseField[i+1].(float64); ok {
						licenses[licName] = int(count)
					}
				}
			}
		}
	}

	return licenses, nil
}

// 解析许可证字符串为LicenseInfo
func parseLicense(licenseStr string) response.LicenseInfo {
	// 简单实现，实际应用中可能需要更复杂的解析逻辑
	licenseType := LicenseType(licenseStr)
	licenseCategory := determineLicenseCategory(licenseType)

	info := response.LicenseInfo{
		Name:     licenseStr,
		Type:     string(licenseType),
		Category: string(licenseCategory),
		URL:      fmt.Sprintf("https://opensource.org/licenses/%s", licenseType),
	}

	return info
}

// 确定许可证类别
func determineLicenseCategory(licenseType LicenseType) LicenseCategory {
	licenseStr := string(licenseType)

	// 根据许可证类型确定类别
	switch {
	case strings.Contains(licenseStr, "GPL"):
		return LicenseCategoryCopyleft
	case strings.Contains(licenseStr, "LGPL"):
		return LicenseCategoryWeakCopyleft
	case strings.Contains(licenseStr, "MIT") ||
		strings.Contains(licenseStr, "Apache") ||
		strings.Contains(licenseStr, "BSD"):
		return LicenseCategoryPermissive
	default:
		// 默认为宽松许可
		return LicenseCategoryPermissive
	}
}

// 查找许可证之间的冲突
func findConflicts(licenses map[response.ArtifactRef][]response.LicenseInfo) []response.LicenseConflict {
	var conflicts []response.LicenseConflict

	// 定义不兼容的许可证组合
	incompatiblePairs := map[string]string{
		string(LicenseTypeGPLv2) + "_" + string(LicenseTypeApache2): "GPL-2.0不兼容Apache-2.0",
		string(LicenseTypeGPLv3) + "_" + string(LicenseTypeCDDL):    "GPL-3.0不兼容CDDL-1.0",
		// 可以添加更多不兼容的许可证组合
	}

	// 检查所有许可证组合
	checkedPairs := make(map[string]bool)

	for _, artifactLicenses := range licenses {
		for _, license1 := range artifactLicenses {
			for _, otherArtifactLicenses := range licenses {
				for _, license2 := range otherArtifactLicenses {
					// 跳过相同的许可证
					if license1.Type == license2.Type {
						continue
					}

					// 创建许可证对的唯一标识符
					pairKey1 := license1.Type + "_" + license2.Type
					pairKey2 := license2.Type + "_" + license1.Type

					// 如果已经检查过这对许可证，则跳过
					if checkedPairs[pairKey1] || checkedPairs[pairKey2] {
						continue
					}

					// 标记为已检查
					checkedPairs[pairKey1] = true
					checkedPairs[pairKey2] = true

					// 检查是否有冲突
					if reason, hasConflict := incompatiblePairs[pairKey1]; hasConflict {
						conflicts = append(conflicts, response.LicenseConflict{
							License1: license1.Type,
							License2: license2.Type,
							Reason:   reason,
						})
					} else if reason, hasConflict := incompatiblePairs[pairKey2]; hasConflict {
						conflicts = append(conflicts, response.LicenseConflict{
							License1: license2.Type,
							License2: license1.Type,
							Reason:   reason,
						})
					}

					// 检查GPL和非GPL许可证的冲突
					if isGPL(license1.Type) && !isGPL(license2.Type) && !isCompatibleWithGPL(license2.Type) {
						conflicts = append(conflicts, response.LicenseConflict{
							License1: license1.Type,
							License2: license2.Type,
							Reason:   fmt.Sprintf("%s不兼容%s", license1.Type, license2.Type),
						})
					}
				}
			}
		}
	}

	return conflicts
}

// 检查是否是GPL许可证
func isGPL(licenseStr string) bool {
	return strings.HasPrefix(licenseStr, "GPL")
}

// 检查许可证是否与GPL兼容
func isCompatibleWithGPL(licenseStr string) bool {
	// 以下许可证通常与GPL兼容
	compatibleLicenses := map[string]bool{
		string(LicenseTypeMIT):       true,
		string(LicenseTypeBSD2):      true,
		string(LicenseTypeBSD3):      true,
		string(LicenseTypeLGPLv2):    true,
		string(LicenseTypeLGPLv3):    true,
		string(LicenseTypeUnlicense): true,
	}

	return compatibleLicenses[licenseStr]
}

// CheckLicenseCompatibility 检查两个许可证是否兼容
func (c *Client) CheckLicenseCompatibility(license1, license2 string) (bool, string, error) {
	// 相同许可证总是兼容的
	if license1 == license2 {
		return true, "相同的许可证总是兼容的", nil
	}

	// 定义不兼容的许可证组合
	incompatiblePairs := map[string]string{
		string(LicenseTypeGPLv2) + "_" + string(LicenseTypeApache2): "GPL-2.0不兼容Apache-2.0",
		string(LicenseTypeGPLv3) + "_" + string(LicenseTypeCDDL):    "GPL-3.0不兼容CDDL-1.0",
		// 这里可以添加更多不兼容的许可证组合
	}

	// 检查是否存在已知的不兼容性
	pairKey1 := license1 + "_" + license2
	pairKey2 := license2 + "_" + license1

	if reason, hasConflict := incompatiblePairs[pairKey1]; hasConflict {
		return false, reason, nil
	}

	if reason, hasConflict := incompatiblePairs[pairKey2]; hasConflict {
		return false, reason, nil
	}

	// 检查GPL和非GPL许可证的冲突
	if isGPL(license1) && !isGPL(license2) && !isCompatibleWithGPL(license2) {
		return false, fmt.Sprintf("%s不兼容%s", license1, license2), nil
	}

	if isGPL(license2) && !isGPL(license1) && !isCompatibleWithGPL(license1) {
		return false, fmt.Sprintf("%s不兼容%s", license2, license1), nil
	}

	// 宽松许可证与其他许可证兼容
	if isPermissiveLicense(license1) || isPermissiveLicense(license2) {
		return true, "宽松许可证通常兼容其他许可证", nil
	}

	// 默认为兼容，但提示不确定
	return true, "未检测到明确的不兼容性，但请咨询法律专家", nil
}

// 检查是否是宽松许可证
func isPermissiveLicense(licenseStr string) bool {
	permissiveLicenses := map[string]bool{
		string(LicenseTypeMIT):       true,
		string(LicenseTypeBSD2):      true,
		string(LicenseTypeBSD3):      true,
		string(LicenseTypeApache2):   true,
		string(LicenseTypeUnlicense): true,
	}

	return permissiveLicenses[licenseStr]
}

// GenerateLicenseReport 为一组组件生成许可证报告
func (c *Client) GenerateLicenseReport(ctx context.Context, artifacts []response.ArtifactRef) (*response.LicenseReport, error) {
	summary, err := c.FindLicenseConflicts(ctx, artifacts)
	if err != nil {
		return nil, fmt.Errorf("生成许可证报告失败: %w", err)
	}

	// 计算合规风险
	var highRiskCount, mediumRiskCount, lowRiskCount int
	for _, conflict := range summary.PotentialConflicts {
		// 简单的风险分析，可以根据需要扩展
		if strings.Contains(conflict.License1, "GPL") || strings.Contains(conflict.License2, "GPL") {
			highRiskCount++
		} else if strings.Contains(conflict.License1, "LGPL") || strings.Contains(conflict.License2, "LGPL") {
			mediumRiskCount++
		} else {
			lowRiskCount++
		}
	}

	// 许可证分类统计
	licenseTypeCount := make(map[string]int)
	for licType, count := range summary.LicenseDistribution {
		licenseTypeCount[string(licType)] = count
	}

	// 所有组件的详细许可证信息
	componentLicenses := make([]response.ComponentLicense, 0, len(artifacts))
	for _, artifact := range artifacts {
		// 获取组件许可证信息
		licenses, err := c.GetComponentLicenses(ctx, artifact.GroupId, artifact.ArtifactId, artifact.Version)
		if err != nil {
			// 如果无法获取许可证信息，添加一个未知许可证的记录
			componentLicenses = append(componentLicenses, response.ComponentLicense{
				GroupId:    artifact.GroupId,
				ArtifactId: artifact.ArtifactId,
				Version:    artifact.Version,
				Licenses:   []response.LicenseInfo{},
				Unknown:    true,
			})
			continue
		}

		componentLicenses = append(componentLicenses, response.ComponentLicense{
			GroupId:    artifact.GroupId,
			ArtifactId: artifact.ArtifactId,
			Version:    artifact.Version,
			Licenses:   licenses,
			Unknown:    false,
		})
	}

	// 创建许可证报告
	report := &response.LicenseReport{
		TotalComponents:   len(artifacts),
		LicenseCount:      len(summary.LicenseDistribution),
		ConflictCount:     len(summary.PotentialConflicts),
		ComponentLicenses: componentLicenses,
		ConflictDetails:   summary.PotentialConflicts,
		RiskAssessment: response.RiskAssessment{
			HighRiskCount:   highRiskCount,
			MediumRiskCount: mediumRiskCount,
			LowRiskCount:    lowRiskCount,
		},
		LicenseDistribution: licenseTypeCount,
		Recommendations:     generateRecommendations(summary),
	}

	return report, nil
}

// FilterByLicenseType 根据许可证类型过滤组件
func (c *Client) FilterByLicenseType(ctx context.Context, artifacts []response.ArtifactRef, allowedTypes []string) ([]response.ArtifactRef, []response.ArtifactRef, error) {
	if len(artifacts) == 0 {
		return []response.ArtifactRef{}, []response.ArtifactRef{}, nil
	}

	// 创建允许的许可证类型集合
	allowedLicenses := make(map[string]bool)
	for _, licType := range allowedTypes {
		allowedLicenses[licType] = true
	}

	var compliant, nonCompliant []response.ArtifactRef

	// 检查每个组件的许可证
	for _, artifact := range artifacts {
		licenses, err := c.GetComponentLicenses(ctx, artifact.GroupId, artifact.ArtifactId, artifact.Version)
		if err != nil {
			// 如果无法获取许可证信息，视为不合规
			nonCompliant = append(nonCompliant, artifact)
			continue
		}

		// 检查是否至少有一个许可证在允许列表中
		hasAllowedLicense := false
		for _, license := range licenses {
			if allowedLicenses[license.Type] {
				hasAllowedLicense = true
				break
			}
		}

		if hasAllowedLicense {
			compliant = append(compliant, artifact)
		} else {
			nonCompliant = append(nonCompliant, artifact)
		}
	}

	return compliant, nonCompliant, nil
}

// 生成许可证使用建议
func generateRecommendations(summary *response.LicenseSummary) []string {
	var recommendations []string

	// 基于冲突数量的建议
	if len(summary.PotentialConflicts) > 5 {
		recommendations = append(recommendations, "存在大量许可证冲突，建议进行详细的法律审查")
	} else if len(summary.PotentialConflicts) > 0 {
		recommendations = append(recommendations, "存在一些许可证冲突，请关注冲突详情")
	} else {
		recommendations = append(recommendations, "未发现明显的许可证冲突，合规性良好")
	}

	// 基于许可证类型的建议
	hasGPL := false
	hasLGPL := false
	for licType := range summary.LicenseDistribution {
		if strings.Contains(licType, "GPL-") && !strings.Contains(licType, "LGPL-") {
			hasGPL = true
		}
		if strings.Contains(licType, "LGPL-") {
			hasLGPL = true
		}
	}

	if hasGPL {
		recommendations = append(recommendations, "项目中包含GPL许可证，如果进行商业分发，需要确保整个项目符合GPL要求")
	}

	if hasLGPL {
		recommendations = append(recommendations, "项目中包含LGPL许可证，需要注意动态链接相关要求")
	}

	// 通用建议
	recommendations = append(recommendations, "定期更新依赖并检查许可证变更")
	recommendations = append(recommendations, "确保所有依赖的许可证条款都被正确遵守")

	return recommendations
}
