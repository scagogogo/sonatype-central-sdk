package request

import "fmt"

const SearchRequestLimitMax = 200

// SearchRequest 表示一次搜索请求
type SearchRequest struct {

	// 从第几条开始返回
	Start int

	// 最多返回200条，默认设置为200
	Limit int

	// 查询参数是啥
	Query *Query

	// Core参数
	Core string

	// 排序字段
	SortField string

	// 排序方向（升序/降序）
	SortAscending bool

	// 是否启用聚合
	FacetEnabled bool

	// 聚合字段
	FacetFields []string

	// 查询键，用于标识批量查询
	QueryKey string

	// 其他自定义参数
	CustomParams map[string]string
}

func NewSearchRequest() *SearchRequest {
	return &SearchRequest{
		Start:        0,
		Limit:        SearchRequestLimitMax,
		Query:        NewQuery(),
		CustomParams: make(map[string]string),
	}
}

func (x *SearchRequest) SetStart(start int) *SearchRequest {
	x.Start = start
	return x
}

func (x *SearchRequest) SetLimit(limit int) *SearchRequest {
	x.Limit = limit
	return x
}

func (x *SearchRequest) SetCore(core string) *SearchRequest {
	x.Core = core
	return x
}

func (x *SearchRequest) SetQuery(query *Query) *SearchRequest {
	x.Query = query
	return x
}

// SetSort 设置排序字段和方向
func (x *SearchRequest) SetSort(field string, ascending bool) *SearchRequest {
	x.SortField = field
	x.SortAscending = ascending
	return x
}

// EnableFacet 启用聚合查询
func (x *SearchRequest) EnableFacet(fields ...string) *SearchRequest {
	x.FacetEnabled = true
	x.FacetFields = fields
	return x
}

// SetQueryKey 设置查询键，用于标识批量查询
func (x *SearchRequest) SetQueryKey(key string) *SearchRequest {
	x.QueryKey = key
	return x
}

// GetQueryKey 获取查询键
func (x *SearchRequest) GetQueryKey() string {
	return x.QueryKey
}

// AddCustomParam 添加自定义参数
func (x *SearchRequest) AddCustomParam(key, value string) *SearchRequest {
	x.CustomParams[key] = value
	return x
}

func (x *SearchRequest) ToRequestParams() string {
	params := fmt.Sprintf("q=%s&rows=%d&wt=json&start=%d", x.Query.ToRequestParamValue(), x.Limit, x.Start)

	// 添加Core参数
	if x.Core != "" {
		params += fmt.Sprintf("&core=%s", x.Core)
	}

	// 添加排序参数
	if x.SortField != "" {
		sortOrder := "asc"
		if !x.SortAscending {
			sortOrder = "desc"
		}
		params += fmt.Sprintf("&sort=%s+%s", x.SortField, sortOrder)
	}

	// 添加聚合参数
	if x.FacetEnabled {
		params += "&facet=true"

		if len(x.FacetFields) > 0 {
			for _, field := range x.FacetFields {
				params += fmt.Sprintf("&facet.field=%s", field)
			}
		}
	}

	// 添加自定义参数
	for key, value := range x.CustomParams {
		params += fmt.Sprintf("&%s=%s", key, value)
	}

	return params
}
