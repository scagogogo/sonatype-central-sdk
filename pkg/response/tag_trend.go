package response

// TagTrend 表示标签的趋势分析结果
type TagTrend struct {
	// 标签名称
	Tag string `json:"tag"`

	// 当前使用该标签的构件数量
	CurrentUsageCount int `json:"currentUsageCount"`

	// 活跃度得分(0-1之间的值)，表示该标签的近期活跃程度
	ActivityScore float64 `json:"activityScore"`

	// 近期更新的项目数量
	RecentUpdatesCount int `json:"recentUpdatesCount"`

	// 趋势指标("上升"、"稳定"或"下降")
	Trend string `json:"trend"`
}
