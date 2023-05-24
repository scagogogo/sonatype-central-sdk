package response

type Artifact struct {
	ID            string   `json:"id"`
	GroupId       string   `json:"g"`
	ArtifactId    string   `json:"a"`
	LatestVersion string   `json:"latestVersion"`
	RepositoryID  string   `json:"repositoryId"`
	Packaging     string   `json:"p"`
	Timestamp     int64    `json:"timestamp"`
	VersionCount  int      `json:"versionCount"`
	Text          []string `json:"text"`
	Ec            []string `json:"ec"`
}
