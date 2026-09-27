package pragoai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	getBoolToolName   = "answer_bool"
	getIntToolName    = "answer_int"
	getFloatToolName  = "answer_float"
	getEnumToolName   = "answer_enum"
	getObjectToolName = "answer_object"

	getStringListToolName = "answer_string_list"
	getBoolListToolName   = "answer_bool_list"
	getIntListToolName    = "answer_int_list"
	getFloatListToolName  = "answer_float_list"
)

func valueDescription(base, responseDescription string) string {
	if responseDescription == "" {
		return base
	}
	return base + " " + responseDescription
}

func stringsToAny(values []string) []any {
	ret := make([]any, len(values))
	for i, v := range values {
		ret[i] = v
	}
	return ret
}

// buildContentBlocks turns request.Text and request.Files into the content
// blocks of the user message. Files must be images or PDFs.
func buildContentBlocks(request Request) ([]anthropic.ContentBlockParamUnion, error) {
	blocks := []anthropic.ContentBlockParamUnion{anthropic.NewTextBlock(request.Text)}

	for _, file := range request.Files {
		encodedData := base64.StdEncoding.EncodeToString(file.Data)
		switch {
		case strings.HasPrefix(file.MediaType, "image/"):
			blocks = append(blocks, anthropic.NewImageBlockBase64(file.MediaType, encodedData))
		case file.MediaType == "application/pdf":
			blocks = append(blocks, anthropic.NewDocumentBlock(anthropic.Base64PDFSourceParam{
				Data: encodedData,
			}))
		default:
			return nil, fmt.Errorf("pragoai: unsupported file media type %q", file.MediaType)
		}
	}

	return blocks, nil
}

// setValueTool registers a single-parameter tool named toolName, forces the
// model to call it, and describes its "value" argument with valueSchema.
func setValueTool(params *anthropic.MessageNewParams, toolName string, valueSchema map[string]any) {
	params.Tools = []anthropic.ToolUnionParam{
		{
			OfTool: &anthropic.ToolParam{
				Name: toolName,
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: map[string]any{
						"value": valueSchema,
					},
					Required: []string{"value"},
				},
			},
		},
	}
	params.ToolChoice = anthropic.ToolChoiceUnionParam{
		OfTool: &anthropic.ToolChoiceToolParam{
			Name: toolName,
		},
	}
}

// ResponseMetadata popisuje spotřebu a výsledek jednoho volání Get /
// GetWithMetadata.
type ResponseMetadata struct {
	Model                    string
	InputTokens              int64
	OutputTokens             int64
	CacheReadInputTokens     int64
	CacheCreationInputTokens int64
	StopReason               string
}

// Get sends request to the AI and returns the answer parsed as T.
// Supported types: string, bool, int, int64, float64, their slice forms
// ([]string, []bool, []int, []int64, []float64), and arbitrary structs
// (including nested structs and slices of the above), whose JSON schema
// is derived from the struct via reflection (see structSchema).
func (client *Client) Get[T any](request Request) (T, error) {
	value, _, err := client.GetWithMetadata[T](request)
	return value, err
}

// GetWithMetadata works like Get, but also returns metadata about the
// call (token usage, stop reason).
func (client *Client) GetWithMetadata[T any](request Request) (T, ResponseMetadata, error) {

	if client == nil {
		var ret T
		var metadata ResponseMetadata
		return ret, metadata, fmt.Errorf("client not defined")
	}

	var zero T
	var metadata ResponseMetadata

	anthropicClient := anthropic.NewClient(
		option.WithAPIKey(client.apiKey),
	)

	model := request.Model
	if model == "" {
		model = DefaultModelLow
	}

	maxTokens := request.MaxTokens
	if maxTokens == 0 {
		maxTokens = DefaultMaxTokens
	}

	contentBlocks, err := buildContentBlocks(request)
	if err != nil {
		return zero, metadata, err
	}

	params := anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: maxTokens,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(contentBlocks...),
		},
	}

	if request.Effort != "" {
		params.OutputConfig.Effort = anthropic.OutputConfigEffort(request.Effort)
	}

	var toolName string

	switch any(zero).(type) {
	case string:
		if len(request.EnumValues) > 0 {
			toolName = getEnumToolName
			setValueTool(&params, toolName, map[string]any{
				"type":        "string",
				"enum":        stringsToAny(request.EnumValues),
				"description": valueDescription("Odpověď na dotaz jako jedna z povolených hodnot.", request.ResponseDescription),
			})
		}
		// jinak plain text answer, no tool needed
	case bool:
		toolName = getBoolToolName
		setValueTool(&params, toolName, map[string]any{
			"type":        "boolean",
			"description": valueDescription("Odpověď na dotaz: true, nebo false.", request.ResponseDescription),
		})
	case int, int64:
		toolName = getIntToolName
		setValueTool(&params, toolName, map[string]any{
			"type":        "integer",
			"description": valueDescription("Odpověď na dotaz jako celé číslo.", request.ResponseDescription),
		})
	case float64:
		toolName = getFloatToolName
		setValueTool(&params, toolName, map[string]any{
			"type":        "number",
			"description": valueDescription("Odpověď na dotaz jako desetinné číslo.", request.ResponseDescription),
		})
	case []string:
		toolName = getStringListToolName
		itemSchema := map[string]any{"type": "string"}
		description := "Odpověď na dotaz jako seznam textových hodnot."
		if len(request.EnumValues) > 0 {
			itemSchema["enum"] = stringsToAny(request.EnumValues)
			description = "Odpověď na dotaz jako seznam hodnot, kde každá je jedna z povolených hodnot."
		}
		setValueTool(&params, toolName, map[string]any{
			"type":        "array",
			"items":       itemSchema,
			"description": valueDescription(description, request.ResponseDescription),
		})
	case []bool:
		toolName = getBoolListToolName
		setValueTool(&params, toolName, map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "boolean"},
			"description": valueDescription("Odpověď na dotaz jako seznam pravdivostních hodnot.", request.ResponseDescription),
		})
	case []int, []int64:
		toolName = getIntListToolName
		setValueTool(&params, toolName, map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "integer"},
			"description": valueDescription("Odpověď na dotaz jako seznam celých čísel.", request.ResponseDescription),
		})
	case []float64:
		toolName = getFloatListToolName
		setValueTool(&params, toolName, map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "number"},
			"description": valueDescription("Odpověď na dotaz jako seznam desetinných čísel.", request.ResponseDescription),
		})
	default:
		t := reflect.TypeOf(zero)
		if t == nil {
			return zero, metadata, fmt.Errorf("pragoai: unsupported type %T", zero)
		}
		schema, err := structSchema(t)
		if err != nil {
			return zero, metadata, fmt.Errorf("pragoai: unsupported type %T: %w", zero, err)
		}
		schema["description"] = valueDescription("Odpověď na dotaz jako objekt.", request.ResponseDescription)
		toolName = getObjectToolName
		setValueTool(&params, toolName, schema)
	}

	response, err := anthropicClient.Messages.New(context.Background(), params)
	if err != nil {
		return zero, metadata, err
	}

	metadata = ResponseMetadata{
		Model:                    string(response.Model),
		InputTokens:              response.Usage.InputTokens,
		OutputTokens:             response.Usage.OutputTokens,
		CacheReadInputTokens:     response.Usage.CacheReadInputTokens,
		CacheCreationInputTokens: response.Usage.CacheCreationInputTokens,
		StopReason:               string(response.StopReason),
	}

	if toolName == "" {
		var result string
		for _, block := range response.Content {
			if textBlock, ok := block.AsAny().(anthropic.TextBlock); ok {
				result += textBlock.Text
			}
		}
		return any(result).(T), metadata, nil
	}

	for _, block := range response.Content {
		toolUseBlock, ok := block.AsAny().(anthropic.ToolUseBlock)
		if !ok || toolUseBlock.Name != toolName {
			continue
		}

		var result struct {
			Value T `json:"value"`
		}
		if err := json.Unmarshal(toolUseBlock.Input, &result); err != nil {
			return zero, metadata, err
		}
		return result.Value, metadata, nil
	}

	return zero, metadata, fmt.Errorf("pragoai: model did not return a %T answer", zero)
}
