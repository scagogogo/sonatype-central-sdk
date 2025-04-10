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
