package response

// TagCount 表示标签及其出现次数
type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}
