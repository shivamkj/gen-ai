package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

const systemPrompt = "You are an AI programming assistant. Don't explain code unless asked. "

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func handleTest(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]string{"message": "OK"})
}

func handleGetAllChats(w http.ResponseWriter, _ *http.Request) {
	rows, err := db.Query(`SELECT c.id, c.model, c.provider, c.title, c.created_at FROM chats c ORDER BY c.created_at DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type chatRow struct {
		ID        int64  `json:"id"`
		Model     string `json:"model"`
		Provider  string `json:"provider"`
		Title     string `json:"title"`
		CreatedAt string `json:"created_at"`
	}
	chats := []chatRow{}
	for rows.Next() {
		var c chatRow
		if err := rows.Scan(&c.ID, &c.Model, &c.Provider, &c.Title, &c.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		chats = append(chats, c)
	}
	writeJSON(w, chats)
}

func handleStartChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Message   string   `json:"message"`
		Model     string   `json:"model"`
		Provider  string   `json:"provider"`
		ImageData []string `json:"imageData"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Message == "" || req.Model == "" || req.Provider == "" {
		http.Error(w, "message, model, and provider are required", http.StatusBadRequest)
		return
	}

	title := generateTitle(req.Message)
	result, err := db.Exec(`INSERT INTO chats (model, provider, title) VALUES (?, ?, ?)`, req.Model, req.Provider, title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	chatID, _ := result.LastInsertId()

	finalMessage := systemPrompt + "\n" + req.Message
	aiResp, err := processChat(chatID, finalMessage, req.Model, req.Provider, req.ImageData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"chat":     map[string]any{"id": chatID, "model": req.Model},
		"response": aiResp,
	})
}

func handleReplyChat(w http.ResponseWriter, r *http.Request) {
	chatID := r.PathValue("id")

	var req struct {
		Message   string   `json:"message"`
		ImageData []string `json:"imageData"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Message == "" && len(req.ImageData) == 0 {
		http.Error(w, "message or image is required", http.StatusBadRequest)
		return
	}

	var model, provider string
	if err := db.QueryRow(`SELECT model, provider FROM chats WHERE id = ?`, chatID).Scan(&model, &provider); err != nil {
		http.Error(w, "chat not found", http.StatusNotFound)
		return
	}

	aiResp, err := processChat(chatID, req.Message, model, provider, req.ImageData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, aiResp)
}

func handleGetMessages(w http.ResponseWriter, r *http.Request) {
	chatID := r.PathValue("id")

	rows, err := db.Query(`
		SELECT m.id, m.role, m.content, m.image_data, m.created_at, m.input_token, m.output_token
		FROM messages m
		WHERE m.chat_id = ?
		ORDER BY m.created_at ASC`, chatID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type msgRow struct {
		ID          int64    `json:"id"`
		Role        string   `json:"role"`
		Content     string   `json:"content"`
		ImageData   []string `json:"image_data"`
		CreatedAt   string   `json:"created_at"`
		InputToken  *int     `json:"input_token"`
		OutputToken *int     `json:"output_token"`
	}
	messages := []msgRow{}
	for rows.Next() {
		var m msgRow
		var imgData *string
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &imgData, &m.CreatedAt, &m.InputToken, &m.OutputToken); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if imgData != nil && *imgData != "" {
			if strings.HasPrefix(*imgData, "[") {
				json.Unmarshal([]byte(*imgData), &m.ImageData)
			} else {
				m.ImageData = []string{*imgData}
			}
		}
		messages = append(messages, m)
	}
	writeJSON(w, messages)
}

func handleDeleteChat(w http.ResponseWriter, r *http.Request) {
	chatID := r.PathValue("id")

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tx.Exec(`DELETE FROM messages WHERE chat_id = ?`, chatID)
	tx.Exec(`DELETE FROM chats WHERE id = ?`, chatID)
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, "OK")
}

func handleDeleteMessage(w http.ResponseWriter, r *http.Request) {
	messageID := r.PathValue("id")
	db.Exec(`DELETE FROM messages WHERE id = ?`, messageID)
	writeJSON(w, "OK")
}
