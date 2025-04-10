package response

type Response[Doc any] struct {
	ResponseHeader *ResponseHeader                `json:"responseHeader"`
	ResponseBody   *ResponseBody[Doc]             `json:"response"`
	FacetCounts    *FacetCounts                   `json:"facet_counts,omitempty"`
	Highlighting   map[string]map[string][]string `json:"highlighting,omitempty"`
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
