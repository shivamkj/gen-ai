package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type Bedrock struct{}

type (
	// ********* Request Data *********
	imageSource struct {
		Type      string `json:"type"`
		MediaType string `json:"media_type"`
		Data      string `json:"data"`
	}
	bedrockContent struct {
		Type   string       `json:"type"`
		Text   string       `json:"text,omitempty"`
		Source *imageSource `json:"source,omitempty"`
	}
	bedrockMsg struct {
		Role    string `json:"role"`
		Content any    `json:"content"`
	}

	// ********* Response Data *********
	ResponseContent struct {
		Text string `json:"text"`
	}

	ResponseUsage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	}

	Response struct {
		Content []ResponseContent `json:"content"`
		Usage   ResponseUsage     `json:"usage"`
	}
)

func (b *Bedrock) Call(model string, messages []Message) (*AiResponse, error) {
	region := os.Getenv("AWS_REGION")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if region == "" || accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("missing AWS credentials: AWS_REGION=%q, AWS_ACCESS_KEY_ID set=%v, AWS_SECRET_ACCESS_KEY set=%v",
			region, accessKey != "", secretKey != "")
	}

	client := bedrockruntime.New(bedrockruntime.Options{
		Region:      region,
		Credentials: credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
	})

	type ()

	bedrockMessages := make([]bedrockMsg, 0, len(messages))
	for _, msg := range messages {
		if len(msg.ImageData) > 0 {
			var parts []bedrockContent
			for _, img := range msg.ImageData {
				mediaType, base64Data := "", ""
				if idx := strings.Index(img, ";base64,"); idx >= 0 {
					mediaType = strings.TrimPrefix(img[:idx], "data:")
					base64Data = img[idx+8:]
				}
				parts = append(parts, bedrockContent{Type: "image", Source: &imageSource{Type: "base64", MediaType: mediaType, Data: base64Data}})
			}
			text := msg.Content
			if text == "" {
				text = "What is in this image?"
			}
			parts = append(parts, bedrockContent{Type: "text", Text: text})
			bedrockMessages = append(bedrockMessages, bedrockMsg{Role: msg.Role, Content: parts})
		} else {
			bedrockMessages = append(bedrockMessages, bedrockMsg{Role: msg.Role, Content: msg.Content})
		}
	}

	payload, _ := json.Marshal(map[string]any{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        20000,
		"messages":          bedrockMessages,
		"temperature":       0.2,
	})

	out, err := client.InvokeModel(context.Background(), &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(model),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        payload,
	})
	if err != nil {
		return nil, fmt.Errorf("bedrock invoke error: %w", err)
	}

	var result Response
	if err := json.Unmarshal(out.Body, &result); err != nil {
		return nil, fmt.Errorf("bedrock decode error: %w", err)
	}

	content := ""
	if len(result.Content) > 0 {
		content = result.Content[0].Text
	}
	in := result.Usage.InputTokens
	outTokens := result.Usage.OutputTokens
	return &AiResponse{Content: content, InputToken: &in, OutputToken: &outTokens}, nil
}
