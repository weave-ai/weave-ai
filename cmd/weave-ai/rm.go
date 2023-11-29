package main

import "github.com/spf13/cobra"

var removeCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove an LLM",
}
