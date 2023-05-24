package api

type Client struct {
	proxy string
}

func NewClient(proxy string) *Client {
	return &Client{}
}
