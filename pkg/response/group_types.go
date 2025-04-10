package response

// GroupArtifact 表示组内的一个工件简要信息
type GroupArtifact struct {
	ArtifactId string `json:"artifactId"`
	Version    string `json:"version"`
}

// GroupSearchResult 表示组搜索结果
type GroupSearchResult struct {
	GroupId         string           `json:"groupId"`
	ArtifactCount   int              `json:"artifactCount"`
	LastUpdated     float64          `json:"lastUpdated"`
	LastUpdatedDate string           `json:"lastUpdatedDate"`
	Artifacts       []*GroupArtifact `json:"artifacts"`
}

// ArtifactStatistics 表示组内一个工件的统计信息
type ArtifactStatistics struct {
	ArtifactId    string `json:"artifactId"`
	VersionCount  int    `json:"versionCount"`
	LatestVersion string `json:"latestVersion"`
}

// GroupStatistics 表示一个组的统计信息
type GroupStatistics struct {
	GroupId         string                `json:"groupId"`
	ArtifactCount   int                   `json:"artifactCount"`
	TotalVersions   int                   `json:"totalVersions"`
	LatestUpdate    int64                 `json:"latestUpdate"`
	LastUpdatedDate string                `json:"lastUpdatedDate"`
	Artifacts       []*ArtifactStatistics `json:"artifacts"`
}

// GroupPopularity 表示组的流行度信息
type GroupPopularity struct {
	GroupId        string `json:"groupId"`
	ArtifactCount  int    `json:"artifactCount"`
	PopularityRank int    `json:"popularityRank"`
}

// GroupComparison 表示两个组的比较结果
type GroupComparison struct {
	Group1              string           `json:"group1"`
	Group2              string           `json:"group2"`
	Group1Stats         *GroupStatistics `json:"group1Stats,omitempty"`
	Group2Stats         *GroupStatistics `json:"group2Stats,omitempty"`
	Group1Error         string           `json:"group1Error,omitempty"`
	Group2Error         string           `json:"group2Error,omitempty"`
	CommonArtifacts     []string         `json:"commonArtifacts,omitempty"`
	CommonArtifactCount int              `json:"commonArtifactCount"`
}

// GroupInfo 表示一个组的基本信息
type GroupInfo struct {
	GroupId         string `json:"groupId"`
	ArtifactCount   int    `json:"artifactCount"`
	LastUpdated     int64  `json:"lastUpdated"`
	LastUpdatedDate string `json:"lastUpdatedDate"`
	Description     string `json:"description,omitempty"`
	Website         string `json:"website,omitempty"`
}
