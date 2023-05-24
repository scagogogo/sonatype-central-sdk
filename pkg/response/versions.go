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
