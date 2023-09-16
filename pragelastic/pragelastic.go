package pragelastic

import (
	"context"
	"time"

	"github.com/olivere/elastic/v7"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
)

type Client struct {
	prefix      string
	esclientOld *elastic.Client
	esclientNew *elasticsearch7.Client
}

func New(id string) (*Client, error) {
	client, err := elastic.NewClient(
		elastic.SetMaxRetries(0),
		elastic.SetHealthcheckTimeoutStartup(100*time.Millisecond),
	)
	if err != nil {
		return nil, err
	}

	esClient, err := elasticsearch7.NewDefaultClient()
	if err != nil {
		return nil, err
	}

	return &Client{
		prefix:      id,
		esclientOld: client,
		esclientNew: esClient,
	}, nil
}

func (c *Client) GetStats() *elastic.ClusterStatsResponse {
	ret, _ := c.esclientOld.ClusterStats().Do(context.Background())
	return ret
}
