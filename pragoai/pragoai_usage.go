package pragoai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const usageReportURL = "https://api.anthropic.com/v1/organizations/usage_report/messages"

type Usage struct {
	Model                    string
	UncachedInputTokens      int64
	CacheReadInputTokens     int64
	CacheCreationInputTokens int64
	OutputTokens             int64
}

type usageReportResponse struct {
	Data []struct {
		Results []struct {
			Model                string `json:"model"`
			UncachedInputTokens  int64  `json:"uncached_input_tokens"`
			CacheReadInputTokens int64  `json:"cache_read_input_tokens"`
			CacheCreation        struct {
				Ephemeral1hInputTokens int64 `json:"ephemeral_1h_input_tokens"`
				Ephemeral5mInputTokens int64 `json:"ephemeral_5m_input_tokens"`
			} `json:"cache_creation"`
			OutputTokens int64 `json:"output_tokens"`
		} `json:"results"`
	} `json:"data"`
	HasMore  bool    `json:"has_more"`
	NextPage *string `json:"next_page"`
}

func (client *Client) GetUsage(startingAt, endingAt time.Time) ([]Usage, error) {
	usageByModel := map[string]*Usage{}

	page := ""
	for {
		values := url.Values{}
		values.Set("starting_at", startingAt.UTC().Format(time.RFC3339))
		values.Set("ending_at", endingAt.UTC().Format(time.RFC3339))
		values.Set("bucket_width", "1d")
		values.Add("group_by[]", "model")
		if page != "" {
			values.Set("page", page)
		}

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, usageReportURL+"?"+values.Encode(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("anthropic-version", "2023-06-01")
		req.Header.Set("x-api-key", client.apiAdminKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("pragoai: usage report request failed: %s: %s", resp.Status, string(body))
		}

		var parsed usageReportResponse
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, err
		}

		for _, bucket := range parsed.Data {
			for _, result := range bucket.Results {
				u, ok := usageByModel[result.Model]
				if !ok {
					u = &Usage{Model: result.Model}
					usageByModel[result.Model] = u
				}
				u.UncachedInputTokens += result.UncachedInputTokens
				u.CacheReadInputTokens += result.CacheReadInputTokens
				u.CacheCreationInputTokens += result.CacheCreation.Ephemeral1hInputTokens + result.CacheCreation.Ephemeral5mInputTokens
				u.OutputTokens += result.OutputTokens
			}
		}

		if !parsed.HasMore || parsed.NextPage == nil {
			break
		}
		page = *parsed.NextPage
	}

	var ret []Usage
	for _, u := range usageByModel {
		ret = append(ret, *u)
	}
	return ret, nil
}
