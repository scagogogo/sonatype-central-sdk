package response

type Response[Doc any] struct {
	ResponseHeader *ResponseHeader    `json:"responseHeader"`
	ResponseBody   *ResponseBody[Doc] `json:"response"`
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
