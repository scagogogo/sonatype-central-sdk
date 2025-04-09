package response

type Version struct {
	ID         string   `json:"id"`
	GroupId    string   `json:"g"`
	ArtifactId string   `json:"a"`
	Version    string   `json:"v"`
	Packaging  string   `json:"p"`
	Timestamp  int64    `json:"timestamp"`
	Ec         []string `json:"ec"`
	Tags       []string `json:"tags"`
}

// VersionInfo 存储版本的详细信息
type VersionInfo struct {
	GroupId     string `json:"groupId"`
	ArtifactId  string `json:"artifactId"`
	Version     string `json:"version"`
	LastUpdated string `json:"lastUpdated"`
	Packaging   string `json:"packaging"`
}

// VersionWithMetadata 包含版本及其元数据信息
type VersionWithMetadata struct {
	Version     *Version     `json:"version"`
	VersionInfo *VersionInfo `json:"versionInfo"`
}

// VersionComparison 比较两个版本的结果
type VersionComparison struct {
	Version1    string `json:"version1"`
	Version2    string `json:"version2"`
	V1Timestamp string `json:"v1Timestamp"`
	V2Timestamp string `json:"v2Timestamp"`
}
