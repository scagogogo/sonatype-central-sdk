package request

import (
	"net/url"
	"strings"
)

// Query 查询参数
type Query struct {

	// 根据组ID搜索
	GroupId string

	// 根据文档ID搜索
	ArtifactId string

	// 根据版本搜索
	Version string

	// 根据标签搜索
	Tags string

	// 根据sha1搜索
	Sha1 string

	// 根据类名来搜索
	ClassName string

	// 根据全路径类名搜索
	FullyQualifiedClassName string

	// 包的类型，比如jar文件
	Packaging string

	Classifier string

	// 自定义查询语句
	CustomQuery string
}

func NewQuery() *Query {
	return &Query{}
}

func (x *Query) SetGroupId(groupId string) *Query {
	x.GroupId = groupId
	return x
}

func (x *Query) SetArtifactId(artifactId string) *Query {
	x.ArtifactId = artifactId
	return x
}

func (x *Query) SetVersion(version string) *Query {
	x.Version = version
	return x
}

func (x *Query) SetTags(tags string) *Query {
	x.Tags = tags
	return x
}

func (x *Query) SetSha1(sha1 string) *Query {
	x.Sha1 = sha1
	return x
}

func (x *Query) SetClassName(className string) *Query {
	x.ClassName = className
	return x
}

func (x *Query) SetFullyQualifiedClassName(fullyQualifiedClassName string) *Query {
	x.FullyQualifiedClassName = fullyQualifiedClassName
	return x
}

func (x *Query) SetPackaging(packaging string) *Query {
	x.Packaging = packaging
	return x
}

func (x *Query) SetClassifier(classifier string) *Query {
	x.Classifier = classifier
	return x
}

// SetCustomQuery 设置自定义查询语句
func (x *Query) SetCustomQuery(query string) *Query {
	x.CustomQuery = query
	return x
}

func (x *Query) ToRequestParamValue() string {
	conditions := make([]string, 0)

	// 如果设置了自定义查询，直接使用自定义查询
	if x.CustomQuery != "" {
		return url.QueryEscape(x.CustomQuery)
	}

	if x.GroupId != "" {
		conditions = append(conditions, "g:"+x.GroupId)
	}

	if x.ArtifactId != "" {
		conditions = append(conditions, "a:"+x.ArtifactId)
	}

	if x.Version != "" {
		conditions = append(conditions, "v:"+x.Version)
	}

	if x.Tags != "" {
		conditions = append(conditions, "tags:"+x.Tags)
	}

	if x.Sha1 != "" {
		conditions = append(conditions, "1:"+x.Sha1)
	}

	if x.ClassName != "" {
		conditions = append(conditions, "c:"+x.ClassName)
	}

	if x.FullyQualifiedClassName != "" {
		conditions = append(conditions, "fc:"+x.FullyQualifiedClassName)
	}

	if x.Packaging != "" {
		conditions = append(conditions, "p:"+x.Packaging)
	}

	if x.Classifier != "" {
		conditions = append(conditions, "l:"+x.Classifier)
	}

	return url.QueryEscape(strings.Join(conditions, " AND "))
}
