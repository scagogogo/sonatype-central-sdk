package response

// SecurityRating 表示组件的安全评分
type SecurityRating struct {
	VulnCount  int      `json:"vulnerabilityCount"` // 漏洞数量
	Severity   string   `json:"maxSeverity"`        // 最高严重性级别: CRITICAL, HIGH, MEDIUM, LOW, NONE
	Score      float64  `json:"score"`              // 安全评分 (0-10, 越高越安全)
	Advisories []string `json:"advisories"`         // 安全建议链接列表
}

// Vulnerability 表示一个漏洞信息
type Vulnerability struct {
	ID          string  `json:"id"`           // 漏洞ID
	CVE         string  `json:"cve"`          // CVE编号
	Title       string  `json:"title"`        // 漏洞标题
	Description string  `json:"description"`  // 漏洞描述
	Severity    string  `json:"severity"`     // 严重性级别: CRITICAL, HIGH, MEDIUM, LOW
	CvssScore   float64 `json:"cvssScore"`    // CVSS评分
	CvssVector  string  `json:"cvssVector"`   // CVSS向量
	Advisory    string  `json:"advisoryLink"` // 安全公告链接
}

// VulnerabilityDetails 包含一个组件的详细漏洞信息
type VulnerabilityDetails struct {
	GroupId         string           `json:"groupId"`
	ArtifactId      string           `json:"artifactId"`
	Version         string           `json:"version"`
	Vulnerabilities []*Vulnerability `json:"vulnerabilities"`
}

// SecurityComparison 保存两个版本的安全比较结果
type SecurityComparison struct {
	GroupId         string          `json:"groupId"`
	ArtifactId      string          `json:"artifactId"`
	Version1        string          `json:"version1"`
	Version2        string          `json:"version2"`
	Rating1         *SecurityRating `json:"rating1"`
	Rating2         *SecurityRating `json:"rating2"`
	SaferVersion    string          `json:"saferVersion"`
	ScoreDifference float64         `json:"scoreDifference"`
}

// ArtifactRef 表示对一个组件的引用
type ArtifactRef struct {
	GroupId    string `json:"groupId"`
	ArtifactId string `json:"artifactId"`
	Version    string `json:"version"`
}

// SecurityScanResult 表示一个组件的批量安全扫描结果
type SecurityScanResult struct {
	GroupId        string          `json:"groupId"`
	ArtifactId     string          `json:"artifactId"`
	Version        string          `json:"version"`
	SecurityRating *SecurityRating `json:"securityRating,omitempty"`
	Error          string          `json:"error,omitempty"`
}

// TimelineEntry 表示漏洞时间线中的一个条目
type TimelineEntry struct {
	Version       string  `json:"version"`
	Timestamp     int64   `json:"timestamp"`     // 发布时间戳
	VulnCount     int     `json:"vulnCount"`     // 漏洞数量
	Severity      string  `json:"severity"`      // 最高严重性
	Score         float64 `json:"score"`         // 安全评分
	Change        string  `json:"change"`        // 变化状态: IMPROVED, DEGRADED, STABLE
	ChangeDetails string  `json:"changeDetails"` // 变化详情
}

// VulnerabilityTimeline 表示一个组件的漏洞时间线
type VulnerabilityTimeline struct {
	GroupId    string           `json:"groupId"`
	ArtifactId string           `json:"artifactId"`
	Entries    []*TimelineEntry `json:"entries"`
}

// ComponentVulnOverview 提供组件的整体漏洞概览信息
type ComponentVulnOverview struct {
	GroupId               string                     `json:"groupId"`
	ArtifactId            string                     `json:"artifactId"`
	TotalVersions         int                        `json:"totalVersions"`
	VulnerableVersions    int                        `json:"vulnerableVersions"`
	LatestVersion         string                     `json:"latestVersion"`
	LatestVulnFreeVersion string                     `json:"latestVulnFreeVersion"`
	VersionRatings        map[string]*SecurityRating `json:"versionRatings"`
	SeverityCounts        map[string]int             `json:"severityCounts"`
}
