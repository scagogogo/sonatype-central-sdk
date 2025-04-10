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

// LicenseInfo 包含组件的许可证信息
type LicenseInfo struct {
	Name        string          `json:"name"`
	Type        LicenseType     `json:"type"`
	Category    LicenseCategory `json:"category"`
	URL         string          `json:"url"`
	Description string          `json:"description,omitempty"`
}

// LicenseConflict 表示两个许可证之间的冲突
type LicenseConflict struct {
	License1 LicenseType `json:"license1"`
	License2 LicenseType `json:"license2"`
	Reason   string      `json:"reason"`
}

// LicenseSummary 包含关于许可证使用情况的摘要
type LicenseSummary struct {
	TotalArtifacts       int                           `json:"totalArtifacts"`
	LicenseDistribution  map[LicenseType]int           `json:"licenseDistribution"`
	CategoryDistribution map[LicenseCategory]int       `json:"categoryDistribution"`
	PotentialConflicts   []LicenseConflict             `json:"potentialConflicts,omitempty"`
	ArtifactsByLicense   map[LicenseType][]ArtifactRef `json:"artifactsByLicense,omitempty"`
}

// ArtifactRef 表示Maven组件的引用
type ArtifactRef struct {
	GroupID    string `json:"groupId"`
	ArtifactID string `json:"artifactId"`
	Version    string `json:"version,omitempty"`
}

// GetComponentLicenses 获取一个组件的许可证信息
func (c *Client) GetComponentLicenses(ctx context.Context, groupID, artifactID, version string) ([]LicenseInfo, error) {
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
	var licenses []LicenseInfo
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
func (c *Client) SearchByLicenseType(ctx context.Context, licenseType LicenseType, limit int) ([]ArtifactRef, error) {
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
	artifacts := make([]ArtifactRef, 0, len(resp.ResponseBody.Docs))
	for _, doc := range resp.ResponseBody.Docs {
		groupID, _ := doc["g"].(string)
		artifactID, _ := doc["a"].(string)
		version, _ := doc["v"].(string)

		artifacts = append(artifacts, ArtifactRef{
			GroupID:    groupID,
			ArtifactID: artifactID,
			Version:    version,
		})
	}

	return artifacts, nil
}

// FindLicenseConflicts 检查组件依赖项中的许可证冲突
func (c *Client) FindLicenseConflicts(ctx context.Context, artifacts []ArtifactRef) (*LicenseSummary, error) {
	if len(artifacts) == 0 {
		return &LicenseSummary{}, nil
	}

	// 保存所有发现的许可证
	foundLicenses := make(map[ArtifactRef][]LicenseInfo)
	licenseDistribution := make(map[LicenseType]int)
	categoryDistribution := make(map[LicenseCategory]int)
	artifactsByLicense := make(map[LicenseType][]ArtifactRef)

	// 获取每个组件的许可证信息
	for _, artifact := range artifacts {
		licenses, err := c.GetComponentLicenses(ctx, artifact.GroupID, artifact.ArtifactID, artifact.Version)
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
				artifactsByLicense[license.Type] = []ArtifactRef{}
			}
			artifactsByLicense[license.Type] = append(artifactsByLicense[license.Type], artifact)
		}
	}

	// 检查许可证冲突
	conflicts := findConflicts(foundLicenses)

	return &LicenseSummary{
		TotalArtifacts:       len(artifacts),
		LicenseDistribution:  licenseDistribution,
		CategoryDistribution: categoryDistribution,
		PotentialConflicts:   conflicts,
		ArtifactsByLicense:   artifactsByLicense,
	}, nil
}

// GetPopularLicenses 获取按使用频率排序的流行许可证
func (c *Client) GetPopularLicenses(ctx context.Context, limit int) (map[LicenseType]int, error) {
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
	licenses := make(map[LicenseType]int)

	if result.FacetCounts != nil && result.FacetCounts.FacetFields != nil {
		if licenseField, ok := result.FacetCounts.FacetFields["l"]; ok {
			// facet结果格式为[license1, count1, license2, count2, ...]
			for i := 0; i < len(licenseField); i += 2 {
				if licName, ok := licenseField[i].(string); ok {
					if count, ok := licenseField[i+1].(float64); ok {
						licenses[LicenseType(licName)] = int(count)
					}
				}
			}
		}
	}

	return licenses, nil
}

// 解析许可证字符串为LicenseInfo
func parseLicense(licenseStr string) LicenseInfo {
	// 简单实现，实际应用中可能需要更复杂的解析逻辑
	licenseType := LicenseType(licenseStr)

	info := LicenseInfo{
		Name: licenseStr,
		Type: licenseType,
	}

	// 根据许可证类型确定类别
	switch {
	case strings.Contains(licenseStr, "GPL"):
		info.Category = LicenseCategoryCopyleft
	case strings.Contains(licenseStr, "LGPL"):
		info.Category = LicenseCategoryWeakCopyleft
	case strings.Contains(licenseStr, "MIT") ||
		strings.Contains(licenseStr, "Apache") ||
		strings.Contains(licenseStr, "BSD"):
		info.Category = LicenseCategoryPermissive
	default:
		// 默认为宽松许可
		info.Category = LicenseCategoryPermissive
	}

	// 设置许可证URL
	info.URL = fmt.Sprintf("https://opensource.org/licenses/%s", licenseType)

	return info
}

// 查找许可证之间的冲突
func findConflicts(licenses map[ArtifactRef][]LicenseInfo) []LicenseConflict {
	var conflicts []LicenseConflict

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
					pairKey1 := string(license1.Type) + "_" + string(license2.Type)
					pairKey2 := string(license2.Type) + "_" + string(license1.Type)

					// 如果已经检查过这对许可证，则跳过
					if checkedPairs[pairKey1] || checkedPairs[pairKey2] {
						continue
					}

					// 标记为已检查
					checkedPairs[pairKey1] = true
					checkedPairs[pairKey2] = true

					// 检查是否有冲突
					if reason, hasConflict := incompatiblePairs[pairKey1]; hasConflict {
						conflicts = append(conflicts, LicenseConflict{
							License1: license1.Type,
							License2: license2.Type,
							Reason:   reason,
						})
					} else if reason, hasConflict := incompatiblePairs[pairKey2]; hasConflict {
						conflicts = append(conflicts, LicenseConflict{
							License1: license2.Type,
							License2: license1.Type,
							Reason:   reason,
						})
					}

					// 检查GPL和非GPL许可证的冲突
					if isGPL(license1.Type) && !isGPL(license2.Type) && !isCompatibleWithGPL(license2.Type) {
						conflicts = append(conflicts, LicenseConflict{
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
func isGPL(license LicenseType) bool {
	return strings.HasPrefix(string(license), "GPL")
}

// 检查许可证是否与GPL兼容
func isCompatibleWithGPL(license LicenseType) bool {
	// 以下许可证通常与GPL兼容
	compatibleLicenses := map[LicenseType]bool{
		LicenseTypeMIT:       true,
		LicenseTypeBSD2:      true,
		LicenseTypeBSD3:      true,
		LicenseTypeLGPLv2:    true,
		LicenseTypeLGPLv3:    true,
		LicenseTypeUnlicense: true,
	}

	return compatibleLicenses[license]
}
