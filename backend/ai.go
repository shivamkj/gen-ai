package main

import (
	"encoding/json"
	"fmt"

	"gen-ai/providers"
)

var providerRegistry map[string]providers.Provider

// processChat fetches chat history, calls the AI provider, and saves both messages to DB.
func processChat(chatID any, message, model, provider string, imageData []string) (*providers.AiResponse, error) {
	rows, err := db.Query(
		`SELECT role, content, image_data FROM messages WHERE chat_id = ? ORDER BY created_at ASC`,
		chatID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []providers.Message
	for rows.Next() {
		var m providers.Message
		var imgData *string
		if err := rows.Scan(&m.Role, &m.Content, &imgData); err != nil {
			return nil, err
		}
		if imgData != nil && *imgData != "" {
			json.Unmarshal([]byte(*imgData), &m.ImageData)
		}
		messages = append(messages, m)
	}
	messages = append(messages, providers.Message{Role: "user", Content: message, ImageData: imageData})

	p, ok := providerRegistry[provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	aiResp, err := p.Call(model, messages)
	if err != nil {
		return nil, err
	}

	var imgVal any
	if len(imageData) > 0 {
		imgJSON, _ := json.Marshal(imageData)
		imgVal = string(imgJSON)
	}
	_, err = db.Exec(
		`INSERT INTO messages (chat_id, role, content, image_data, input_token, output_token)
		 VALUES (?, ?, ?, ?, NULL, NULL),
		        (?, ?, ?, NULL, ?, ?)`,
		chatID, "user", message, imgVal,
		chatID, "assistant", aiResp.Content, aiResp.InputToken, aiResp.OutputToken,
	)
	if err != nil {
		return nil, err
	}
	return aiResp, nil
}
