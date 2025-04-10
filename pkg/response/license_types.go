package response

// LicenseInfo 包含组件的许可证信息
type LicenseInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// LicenseConflict 表示两个许可证之间的冲突
type LicenseConflict struct {
	License1 string `json:"license1"`
	License2 string `json:"license2"`
	Reason   string `json:"reason"`
}

// LicenseSummary 包含关于许可证使用情况的摘要
type LicenseSummary struct {
	TotalArtifacts       int                      `json:"totalArtifacts"`
	LicenseDistribution  map[string]int           `json:"licenseDistribution"`
	CategoryDistribution map[string]int           `json:"categoryDistribution"`
	PotentialConflicts   []LicenseConflict        `json:"potentialConflicts,omitempty"`
	ArtifactsByLicense   map[string][]ArtifactRef `json:"artifactsByLicense,omitempty"`
}

// ComponentLicense 包含组件的许可证详细信息
type ComponentLicense struct {
	GroupId    string        `json:"groupId"`
	ArtifactId string        `json:"artifactId"`
	Version    string        `json:"version"`
	Licenses   []LicenseInfo `json:"licenses"`
	Unknown    bool          `json:"unknown"` // 是否无法确定许可证
}

// RiskAssessment 许可证合规风险评估
type RiskAssessment struct {
	HighRiskCount   int `json:"highRiskCount"`   // 高风险冲突数量
	MediumRiskCount int `json:"mediumRiskCount"` // 中风险冲突数量
	LowRiskCount    int `json:"lowRiskCount"`    // 低风险冲突数量
}

// LicenseReport 许可证合规报告
type LicenseReport struct {
	TotalComponents     int                `json:"totalComponents"`
	LicenseCount        int                `json:"licenseCount"`
	ConflictCount       int                `json:"conflictCount"`
	ComponentLicenses   []ComponentLicense `json:"componentLicenses"`
	ConflictDetails     []LicenseConflict  `json:"conflictDetails"`
	RiskAssessment      RiskAssessment     `json:"riskAssessment"`
	LicenseDistribution map[string]int     `json:"licenseDistribution"`
	Recommendations     []string           `json:"recommendations"`
}
