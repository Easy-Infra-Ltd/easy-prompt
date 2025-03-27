package render

import (
	"bufio"
	"fmt"
	"os"

	"github.com/Easy-Infra-Ltd/assert"
	"github.com/Easy-Infra-Ltd/easy-prompt/src/terminal"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type TerminalChat struct {
	writer   *terminal.Writer
	reader   *bufio.Reader
	messages []*message
}

type message struct {
	message string
	author  string
}

func New(messages []*message) *TerminalChat {
	if messages == nil {
		messages = make([]*message, 0)
	}

	return &TerminalChat{
		writer:   terminal.New(os.Stdout),
		messages: messages,
	}
}

func (tc *TerminalChat) RenderMessageAuthor(author string) error {
	assert.Assert(author != "", "Author should never be an empty string, something has gone wrong")

	caser := cases.Title(language.English)
	if _, err := tc.writer.Write([]byte(caser.String(fmt.Sprintf("%s:\n", author)))); err != nil {
		return err
	}

	alignment := terminal.AlignLeft
	if author == "assistant" {
		alignment = terminal.AlignCenter
	}

	if err := tc.writer.Print(alignment); err != nil {
		return err
	}

	tc.writer.Reset()

	return nil
}

func (tc *TerminalChat) RenderMessage(text string, author string) error {
	assert.Assert(text != "", "Text should never be an empty string, something has gone wrong")
	assert.Assert(author != "", "Author should never be an empty string, something has gone wrong")

	m := &message{
		message: text,
		author:  author,
	}

	tc.messages = append(tc.messages, m)

	alignment := terminal.AlignLeft
	if author == "assistant" {
		alignment = terminal.AlignCenter
	}

	if _, err := tc.writer.Write([]byte(text)); err != nil {
		return err
	}

	if err := tc.writer.Print(alignment); err != nil {
		return err
	}

	return nil
}

func (tc *TerminalChat) EndMessage() {
	tc.writer.Reset()
}

func (tc *TerminalChat) ClearMessage() {
	tc.writer.Clear()
}
