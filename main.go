package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	// TODO: Configure APIs and load configuration from ENV
	// TODO:
	var name string
	var rootCmd = &cobra.Command{
		Use:   "ask",
		Short: "A simple CLI app to ask a question and read input",
		Long:  "Usage: ask - A command-line tool to ask a question and capture user input.",
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("What is your favorite programming language? ")
			answer, _ := reader.ReadString('\n')
			fmt.Printf("You answered: %s", answer)
		},
	}

	rootCmd.Flags().StringVarP(&name, "name", "n", "User", "Your name")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
