package main

import "github.com/spf13/cobra"

var createLmCmd = &cobra.Command{
	Use:     "create-lm",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"create-language-model", "create-llm"},
	Short:   "Create and deploy a language model",
	RunE:    createLmCmdRun,
}

var createLmFlags struct {
	model          string
	modelNamespace string
	cpu            string
}

func init() {
	createLmCmd.Flags().StringVarP(&createLmFlags.model, "model", "m", "zephyr-7b-beta", "model name")
	createLmCmd.Flags().StringVarP(&createLmFlags.modelNamespace, "model-ns", "N", *kubeconfigArgs.Namespace, "model namespace")
	createLmCmd.Flags().StringVarP(&createLmFlags.cpu, "cpu", "c", "4", "cpu")

	rootCmd.AddCommand(createLmCmd)
}

func createLmCmdRun(cmd *cobra.Command, args []string) error {
	// lmName := args[0]
	return nil
}
