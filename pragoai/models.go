package pragoai

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func (client *Client) GetModels() ([]Model, error) {
	anthropicClient := anthropic.NewClient(
		option.WithAPIKey(client.apiKey),
	)

	var models []anthropic.ModelInfo
	pager := anthropicClient.Models.ListAutoPaging(context.Background(), anthropic.ModelListParams{})
	for pager.Next() {
		models = append(models, pager.Current())
	}
	if err := pager.Err(); err != nil {
		return nil, err
	}

	var ret []Model

	for _, v := range models {
		ret = append(ret, Model{
			ID:   v.ID,
			Name: v.DisplayName,
		})
	}

	return ret, nil
}
