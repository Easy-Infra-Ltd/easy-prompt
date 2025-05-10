package main

import (
	"fmt"
	"os"

	"github.com/Easy-Infra-Ltd/easy-prompt/src/anthropic"
	"github.com/Easy-Infra-Ltd/easy-prompt/src/render"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "easyprompt",
		Short: "A simple CLI app to ask a question and read input",
		Long:  "Usage: ask - A command-line tool to ask a question and capture user input.",
		RunE: func(cmd *cobra.Command, args []string) error {

			client := anthropic.NewAnthropicClient("")
			renderer := render.New(nil)

			prompt, _ := renderer.GetInput()
			if err := client.StartChat(renderer, anthropic.Claude3HaikuLatest, "", prompt); err != nil {
				return err
			}
			defer client.EndChat()

			for {
				prompt, _ := renderer.GetInput()

				if prompt == "exit\n" {
					break
				}

				message := &anthropic.Message{Role: anthropic.RoleUser, Content: prompt}
				if err := client.SendMessage(message); err != nil {
					return err
				}
			}

			return nil
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
