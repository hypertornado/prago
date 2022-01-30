package pragelastic

import (
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
