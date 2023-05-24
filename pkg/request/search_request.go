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

	Core string
}

func NewSearchRequest() *SearchRequest {
	return &SearchRequest{
		Start: 0,
		Limit: SearchRequestLimitMax,
		Query: NewQuery(),
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

func (x *SearchRequest) ToRequestParams() string {
	return fmt.Sprintf("q=%s&rows=%d&wt=json&start=%d&core=%s", x.Query.ToRequestParamValue(), x.Limit, x.Start, x.Core)
}
