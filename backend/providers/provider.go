package providers

// Message represents a chat message sent to AI APIs.
type Message struct {
	Role      string   `json:"role"`
	Content   string   `json:"content"`
	ImageData []string `json:"image_data,omitempty"`
}

// AiResponse is the unified response from any AI provider.
type AiResponse struct {
	Content     string `json:"content"`
	InputToken  *int   `json:"input_token"`
	OutputToken *int   `json:"output_token"`
}

// Provider is the interface each AI provider must implement.
type Provider interface {
	Call(model string, messages []Message) (*AiResponse, error)
}
