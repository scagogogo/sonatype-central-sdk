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
