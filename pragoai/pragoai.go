package pragoai

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
)

type Client struct {
	provider string
	apiKey   string
}

func NewClaudeClient(apiKey string) *Client {
	return &Client{
		provider: "claude",
		apiKey:   apiKey,
	}
}

type Request struct {
	Text  string
	Model string
}

type Model struct {
	ID   string
	Name string
}

const (
	DefaultModelLow    string = "claude-haiku-4-5-20251001"
	DefaultModelMiddle string = "claude-sonnet-5"
	DefaultModelHigh   string = "claude-opus-5"
)

func (client *Client) GetText(request Request) (string, error) {
	anthropicClient := anthropic.NewClient(
		option.WithAPIKey(client.apiKey),
	)

	model := request.Model
	if model == "" {
		model = DefaultModelLow
	}

	response, err := anthropicClient.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: 16000,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(request.Text)),
		},
	})
	if err != nil {
		return "", err
	}

	var result string
	for _, block := range response.Content {
		if textBlock, ok := block.AsAny().(anthropic.TextBlock); ok {
			result += textBlock.Text
		}
	}

	return result, nil
}

const getBoolToolName = "answer_bool"

func (client *Client) GetBool(request Request) (bool, error) {
	anthropicClient := anthropic.NewClient(
		option.WithAPIKey(client.apiKey),
	)

	model := request.Model
	if model == "" {
		model = DefaultModelLow
	}

	response, err := anthropicClient.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(request.Text)),
		},
		Tools: []anthropic.ToolUnionParam{
			{
				OfTool: &anthropic.ToolParam{
					Name:        getBoolToolName,
					Description: param.NewOpt("Vrátí odpověď na dotaz jako pravdivostní hodnotu."),
					InputSchema: anthropic.ToolInputSchemaParam{
						Properties: map[string]any{
							"value": map[string]any{
								"type":        "boolean",
								"description": "Odpověď na dotaz: true, nebo false.",
							},
						},
						Required: []string{"value"},
					},
				},
			},
		},
		ToolChoice: anthropic.ToolChoiceUnionParam{
			OfTool: &anthropic.ToolChoiceToolParam{
				Name: getBoolToolName,
			},
		},
	})
	if err != nil {
		return false, err
	}

	for _, block := range response.Content {
		toolUseBlock, ok := block.AsAny().(anthropic.ToolUseBlock)
		if !ok || toolUseBlock.Name != getBoolToolName {
			continue
		}

		var result struct {
			Value bool `json:"value"`
		}
		if err := json.Unmarshal(toolUseBlock.Input, &result); err != nil {
			return false, err
		}
		return result.Value, nil
	}

	return false, errors.New("pragoai: model did not return a bool answer")
}
