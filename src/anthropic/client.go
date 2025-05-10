package anthropic

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Easy-Infra-Ltd/assert"
	"modernc.org/sqlite"
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
	assert.Assert(model != "", "Model must have a value")
	assert.Assert(message != "", "Message must have a value")

	c.chat = &AnthropicChat{
		model:        model,
		systemPrompt: systemPrompt,
		messages:     []*Message{},
		renderer:     renderer,
	}

	return c.SendMessage(&Message{Content: message, Role: RoleUser})
}

func (c *AnthropicClient) SendMessage(message *Message) error {
	assert.NotNil(message, "Message passed to SendMessage should not be empty")
	assert.NotNil(c.chat, "Chat must be initialised before sending a message")

	messages := append(c.chat.messages, message)
	requestBody := MessageRequest{
		Model:     c.chat.model,
		Messages:  messages,
		MaxTokens: 1024,
		Thinking: &Think{
			Type: "disabled",
		},
	}

	if c.chat.renderer != nil {
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
	}

	if c.chat.systemPrompt != "" {
		requestBody.System = c.chat.systemPrompt
	}

	assert.NotNil(requestBody, "Request body must not be nil")
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.key)
	req.Header.Set("anthropic-version", ApiVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
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
		return err
	}

	if response.Stop_reason == StopMaxTokens {
		return fmt.Errorf(response.Stop_reason)
	}

	if len(response.Content) == 0 {
		return fmt.Errorf("response content is empty")
	}

	if c.chat.renderer != nil {
		c.chat.renderer.ClearMessage()
	}

	for _, content := range response.Content {
		switch content.Type {
		case TextContent:
			messages = append(messages, &Message{
				Content: content.Text,
				Role:    RoleAssistant,
			})

			if c.chat.renderer != nil {
				if err := c.chat.renderer.RenderMessage(content.Text, RoleAssistant); err != nil {
					return err
				}
			}
		case ImageContent:
			if c.chat.renderer != nil {
				if err := c.chat.renderer.RenderImage(content.Image); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unknown content type: %s", content.Type)
		}
	}

	if c.chat.renderer != nil {
		c.chat.renderer.EndMessage()
	}

	sqlDB, err := sql.Open("sqlite3", "prompt.db")
	c.chat.save(sqlDB)

	c.chat.messages = messages

	return nil
}

func (c *AnthropicClient) EndChat() {
	// Clear down any required parts
	c.chat = nil
}

func (c *AnthropicChat) save(db *sql.DB) error {
	// TODO: use sqlc to generate the insert statement
	_, err := db.Exec("INSERT INTO chats (model, system_prompt, messages) VALUES (?, ?, ?)", c.model, c.systemPrompt, c.messages)
	return err
}
