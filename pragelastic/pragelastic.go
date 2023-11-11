package pragelastic

import (
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
)

type Client struct {
	prefix      string
	esclientNew *elasticsearch7.Client
}

func New(id string) (*Client, error) {

	esClient, err := elasticsearch7.NewClient(elasticsearch7.Config{})
	if err != nil {
		return nil, err
	}

	return &Client{
		prefix:      id,
		esclientNew: esClient,
	}, nil
}

func (client *Client) DeleteIndex(name string) error {
	_, err := client.esclientNew.Indices.Delete([]string{name})
	return err
}
