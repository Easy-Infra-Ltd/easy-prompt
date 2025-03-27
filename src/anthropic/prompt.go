package anthropic

type MessageRequest struct {
	Model       string     `json:"model"`
	MaxTokens   int        `json:"max_tokens,omitempty"`
	Messages    []*Message `json:"messages"`
	Stream      bool       `json:"stream,omitempty"`
	System      string     `json:"system,omitempty"`
	Temperature float32    `json:"temperature,omitempty"`
	Thinking    *Think     `json:"thinking,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func ContentToMessage(content *Content) *Message {
	return &Message{
		Role:    "assistant",
		Content: content.Text,
	}
}

type Think struct {
	Budget_tokens int    `json:"budget_tokens,omitempty"`
	Type          string `json:"type"`
}

type ChatResponse struct {
	Id            string    `json:"id"`
	Content       []Content `json:"content"`
	Model         string    `json:"model"`
	Role          string    `json:"role"`
	Stop_reason   string    `json:"stop_reason"`
	Stop_sequence string    `json:"stop_sequence"`
	Type          string    `json:"type"`
	Usage         Usage     `json:"usage"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"Text"`
}

type Usage struct {
	Input_tokens  int `json:"input_tokens"`
	Output_tokens int `json:"output_tokens"`
}
