package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/Easy-Infra-Ltd/easy-prompt/src/anthropic"
	"github.com/Easy-Infra-Ltd/easy-prompt/src/render"
	"github.com/spf13/cobra"
)

func main() {
	// TODO: Configure APIs and load configuration from ENV
	var rootCmd = &cobra.Command{
		Use:   "easyprompt",
		Short: "A simple CLI app to ask a question and read input",
		Long:  "Usage: ask - A command-line tool to ask a question and capture user input.",
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("What would you like to know today?\n")
			prompt, _ := reader.ReadString('\n')

			client := anthropic.NewAnthropicClient("")
			renderer := render.New(nil)
			err := client.StartChat(renderer, anthropic.Claude3HaikuLatest, "", prompt)
			if err != nil {
				return err
			}

			return nil
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
