package response

// ArtifactMetadata 制品的完整元数据信息
type ArtifactMetadata struct {
	// 基本信息
	GroupId       string `json:"groupId"`
	ArtifactId    string `json:"artifactId"`
	LatestVersion string `json:"latestVersion"`
	Packaging     string `json:"packaging"`
	LastUpdated   int64  `json:"lastUpdated"`

	// POM文件内容
	PomContent string `json:"pomContent,omitempty"`

	// 依赖项
	Dependencies []*Dependency `json:"dependencies,omitempty"`

	// 安全评分
	SecurityRating *SecurityRating `json:"securityRating,omitempty"`

	// 许可证信息
	Licenses []string `json:"licenses,omitempty"`

	// 开发者信息
	Developers []*Developer `json:"developers,omitempty"`

	// 项目信息
	ProjectInfo *ProjectInfo `json:"projectInfo,omitempty"`
}

// Dependency 依赖信息
type Dependency struct {
	GroupId    string `json:"groupId"`
	ArtifactId string `json:"artifactId"`
	Version    string `json:"version"`
	Scope      string `json:"scope,omitempty"`
	Optional   bool   `json:"optional,omitempty"`
}

// SecurityRating 安全评分信息
type SecurityRating struct {
	Score       float64           `json:"score,omitempty"`
	VulnCount   int               `json:"vulnCount,omitempty"`
	Severity    string            `json:"severity,omitempty"`
	Advisories  []string          `json:"advisories,omitempty"`
	Description string            `json:"description,omitempty"`
	Details     map[string]string `json:"details,omitempty"`
}

// Developer 开发者信息
type Developer struct {
	Name    string `json:"name,omitempty"`
	Email   string `json:"email,omitempty"`
	URL     string `json:"url,omitempty"`
	ID      string `json:"id,omitempty"`
	Company string `json:"company,omitempty"`
}

// ProjectInfo 项目信息
type ProjectInfo struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	SCM         string `json:"scm,omitempty"`
	Issues      string `json:"issues,omitempty"`
}

// FacetResults 聚合查询结果
type FacetResults struct {
	Counts map[string][]FacetCount `json:"facet_counts,omitempty"`
}

// FacetCount 聚合计数
type FacetCount struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}
