package request

// AdvancedSearchOptions 高级搜索选项
type AdvancedSearchOptions struct {
	// 组ID
	GroupId string

	// 制品ID
	ArtifactId string

	// 版本
	Version string

	// 包的类型
	Packaging string

	// 分类器
	Classifier string
}

// NewAdvancedSearchOptions 创建新的高级搜索选项
func NewAdvancedSearchOptions() *AdvancedSearchOptions {
	return &AdvancedSearchOptions{}
}

// SetGroupId 设置组ID
func (x *AdvancedSearchOptions) SetGroupId(groupId string) *AdvancedSearchOptions {
	x.GroupId = groupId
	return x
}

// SetArtifactId 设置制品ID
func (x *AdvancedSearchOptions) SetArtifactId(artifactId string) *AdvancedSearchOptions {
	x.ArtifactId = artifactId
	return x
}

// SetVersion 设置版本
func (x *AdvancedSearchOptions) SetVersion(version string) *AdvancedSearchOptions {
	x.Version = version
	return x
}

// SetPackaging 设置包类型
func (x *AdvancedSearchOptions) SetPackaging(packaging string) *AdvancedSearchOptions {
	x.Packaging = packaging
	return x
}

// SetClassifier 设置分类器
func (x *AdvancedSearchOptions) SetClassifier(classifier string) *AdvancedSearchOptions {
	x.Classifier = classifier
	return x
}

// MakeDependencyQuery 创建依赖查询字符串
func MakeDependencyQuery(groupId, artifactId string) string {
	if groupId != "" && artifactId != "" {
		return "d:" + groupId + ":" + artifactId
	} else if groupId != "" {
		return "d:" + groupId
	} else if artifactId != "" {
		return "d:*:" + artifactId
	}
	return ""
}

// MakeLicenseQuery 创建许可证查询字符串
func MakeLicenseQuery(license string) string {
	return "l:" + license
}
