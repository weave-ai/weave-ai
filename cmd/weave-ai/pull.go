package main

import "github.com/spf13/cobra"

var pullCmd = &cobra.Command{
	Use:   "pull",
	Args:  cobra.ExactArgs(1),
	Short: "Pull a model OCI",
	Long: `Pull a model OCI.
# Pull a model OCI.
weave-ai pull flux-7b

# Pull a model OCI.
weave-ai pull ghcr.io/weave-ai/flux-7b:v0.1.0-q5km-gguf
`,
	RunE: pullCmdRun,
}

func init() {
	// rootCmd.AddCommand(pullCmd)
}

func pullCmdRun(cmd *cobra.Command, args []string) error {
	return nil
}
