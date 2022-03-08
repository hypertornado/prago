package pragelastic

import (
	"context"

	"github.com/olivere/elastic/v7"
)

type Client struct {
	prefix  string
	eclient *elastic.Client
}

func New(id string) (*Client, error) {
	client, err := elastic.NewClient()
	if err != nil {
		return nil, err
	}
	return &Client{
		prefix:  id,
		eclient: client,
	}, nil
}

func (c *Client) GetStats() *elastic.ClusterStatsResponse {
	ret, _ := c.eclient.ClusterStats().Do(context.Background())
	return ret
}
