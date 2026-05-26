package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAI is a generic provider for any OpenAI-compatible API.
type OpenAI struct {
	BaseURL string
	APIKey  string
}

type (
	// ********* Request Data *********
	openAiContent struct {
		Type     string `json:"type"`
		Text     string `json:"text,omitempty"`
		ImageURL *struct {
			URL string `json:"url"`
		} `json:"image_url,omitempty"`
	}

	oaiMessage struct {
		Role    string `json:"role"`
		Content any    `json:"content"`
	}

	// ********* Response Data *********
	Choice struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	}

	APIError struct {
		Message string `json:"message"`
	}

	Result struct {
		Choices []Choice  `json:"choices"`
		Usage   Usage     `json:"usage"`
		Error   *APIError `json:"error"`
	}
)

func (o *OpenAI) Call(model string, messages []Message) (*AiResponse, error) {
	oaiMessages := make([]oaiMessage, 0, len(messages))
	for _, msg := range messages {
		if len(msg.ImageData) > 0 {
			parts := []openAiContent{{Type: "text", Text: msg.Content}}
			for _, img := range msg.ImageData {
				parts = append(parts, openAiContent{Type: "image_url", ImageURL: &struct {
					URL string `json:"url"`
				}{URL: img}})
			}
			oaiMessages = append(oaiMessages, oaiMessage{Role: msg.Role, Content: parts})
		} else {
			oaiMessages = append(oaiMessages, oaiMessage{Role: msg.Role, Content: msg.Content})
		}
	}

	body, _ := json.Marshal(map[string]any{
		"model":       model,
		"messages":    oaiMessages,
		"temperature": 0,
		"max_tokens":  8192,
	})

	req, err := http.NewRequest("POST", o.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result Result
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode, string(rawBody))
	}
	if result.Error != nil {
		return nil, fmt.Errorf("openai error: %s", result.Error.Message)
	}

	content := ""
	if len(result.Choices) > 0 {
		content = result.Choices[0].Message.Content
	}
	in := result.Usage.PromptTokens
	out := result.Usage.CompletionTokens
	return &AiResponse{Content: content, InputToken: &in, OutputToken: &out}, nil
}
