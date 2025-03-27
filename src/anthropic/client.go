package anthropic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Easy-Infra-Ltd/assert"
)

type AnthropicClient struct {
	key        string
	baseURL    string
	httpClient *http.Client
	chat       *AnthropicChat
}

type ErrorResponse struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewAnthropicClient(key string) *AnthropicClient {
	if key == "" {
		key = os.Getenv("ANTHROPIC_API_KEY")
		assert.Assert(key != "", "API Key must have a value")
	}

	return &AnthropicClient{
		key:        key,
		baseURL:    "https://api.anthropic.com/v1/messages",
		httpClient: &http.Client{},
	}
}

type AnthropicChat struct {
	messages     []*Message
	model        string
	systemPrompt string
	renderer     ChatRenderer
}

type ChatRenderer interface {
	RenderMessageAuthor(string) error
	RenderMessage(string, string) error
	ClearMessage()
	EndMessage()
}

func (c *AnthropicClient) StartChat(renderer ChatRenderer, model string, systemPrompt string, message string) error {
	c.chat = &AnthropicChat{
		model:        model,
		systemPrompt: systemPrompt,
		messages:     []*Message{},
		renderer:     renderer,
	}

	return c.SendMessage(&Message{Content: message, Role: RoleUser})
}

func (c *AnthropicClient) SendMessage(message *Message) error {
	assert.NotNil(message, "Message passed to SendMessage shoud not be empty")

	messages := append(c.chat.messages, message)
	requestBody := MessageRequest{
		Model:     c.chat.model,
		Messages:  messages,
		MaxTokens: 1024,
		Thinking: &Think{
			Type: "disabled",
		},
	}

	if err := c.chat.renderer.RenderMessageAuthor(message.Role); err != nil {
		return err
	}

	if err := c.chat.renderer.RenderMessage(message.Content, message.Role); err != nil {
		return err
	}

	c.chat.renderer.EndMessage()

	if err := c.chat.renderer.RenderMessageAuthor(RoleAssistant); err != nil {
		return err
	}

	if err := c.chat.renderer.RenderMessage(fmt.Sprintf("%s Thinking...", RoleAssistant), RoleAssistant); err != nil {
		return err
	}

	if c.chat.systemPrompt != "" {
		requestBody.System = c.chat.systemPrompt
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		// TODO: Handle return
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		// TODO: handle return
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.key)
	req.Header.Set("anthropic-version", ApiVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// TODO: handle return
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// TODO: handle return
		return err
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return fmt.Errorf("error status code: %d\nraw response: %s", resp.StatusCode, body)
		}

		return fmt.Errorf("API error: %s - %s", errorResp.Error.Type, errorResp.Error.Message)
	}

	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		// TODO: handle marshal
		return err
	}

	if response.Stop_reason == StopMaxTokens {
		// TODO: handle max tokens
		return err
	}

	if len(response.Content) == 0 {
		// TODO: handle empty content response
		return err
	}

	c.chat.renderer.ClearMessage()

	for _, content := range response.Content {
		if content.Type == TextContent {
			messages = append(messages, &Message{
				Content: content.Text,
				Role:    RoleAssistant,
			})

			if err := c.chat.renderer.RenderMessage(content.Text, RoleAssistant); err != nil {
				return err
			}
		}
	}

	c.chat.renderer.EndMessage()

	// TODO: Update chat to database

	c.chat.messages = messages

	return nil
}
