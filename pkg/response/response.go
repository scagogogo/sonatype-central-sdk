package response

type Response[Doc any] struct {
	ResponseHeader *ResponseHeader    `json:"responseHeader"`
	ResponseBody   *ResponseBody[Doc] `json:"response"`
	FacetCounts    *FacetCounts       `json:"facet_counts,omitempty"`

	// Highlighting 包含搜索结果的高亮信息
	// 数据结构为三层嵌套映射：
	// 1. 第一层key: 文档ID (如 "org.apache:commons-io:1.2.3")
	// 2. 第二层key: 高亮字段名 (如 "fch"表示fully qualified class name)
	// 3. 值: 高亮片段数组，其中匹配部分用<em>标签包围
	//
	// 示例:
	// {
	//   "org.apache:commons-io:1.2.3": {
	//     "fch": ["<em>org.apache</em>.commons.io.FileUtils"]
	//   }
	// }
	Highlighting map[string]map[string][]string `json:"highlighting,omitempty"`
}

type ResponseHeader struct {
	Status int     `json:"status"`
	QTime  int     `json:"QTime"`
	Params *Params `json:"params"`
}

type Params struct {
	Q       string `json:"q"`
	Core    string `json:"core"`
	Indent  string `json:"indent"`
	Fl      string `json:"fl"`
	Start   string `json:"start"`
	Sort    string `json:"sort"`
	Rows    string `json:"rows"`
	Wt      string `json:"wt"`
	Version string `json:"version"`
}

type ResponseBody[Doc any] struct {
	NumFound int   `json:"numFound"`
	Start    int   `json:"start"`
	Docs     []Doc `json:"docs"`
}

// FacetCounts 表示聚合查询结果
type FacetCounts struct {
	FacetFields  map[string][]interface{} `json:"facet_fields,omitempty"`
	FacetQueries map[string]int           `json:"facet_queries,omitempty"`
	FacetDates   map[string]interface{}   `json:"facet_dates,omitempty"`
}
