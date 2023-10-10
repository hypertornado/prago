package pragelastic

import (
	"encoding/json"
	"io"
)

type Stats struct {
	Shards  StatsShards             `json:"_shards"`
	Indices map[string]*StatsIndice `json:"indices"`
}

type StatsShards struct {
	Total      int64 `json:"total"`
	Successful int64 `json:"successful"`
	Failed     int64 `json:"failed"`
}

type StatsIndice struct {
	UUID  string           `json:"uuid"`
	Total StatsIndiceTotal `json:"total"`
}

type StatsIndiceTotal struct {
	Docs  StatsIndiceTotalDocs  `json:"docs"`
	Store StatsIndiceTotalStore `json:"store"`
}

type StatsIndiceTotalDocs struct {
	Count   int64 `json:"count"`
	Deleted int64 `json:"deleted"`
}

type StatsIndiceTotalStore struct {
	SizeInBytes int64 `json:"size_in_bytes"`
}

func (c *Client) GetStats() (*Stats, error) {
	res, err := c.esclientNew.Indices.Stats()
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var ret Stats
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil

}
