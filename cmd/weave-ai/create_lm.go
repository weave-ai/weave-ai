package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/fluxcd/pkg/ssa"
	"github.com/spf13/cobra"
	"github.com/weave-ai/weave-ai/pkg/utils"
)

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
	serviceType    string
	cpu            string
	export         bool
	wait           bool
}

func init() {
	createLmCmd.Flags().StringVarP(&createLmFlags.model, "model", "m", "zephyr-7b-beta", "model name")
	createLmCmd.Flags().StringVarP(&createLmFlags.modelNamespace, "model-ns", "N", defaultNamespace, "model namespace")
	createLmCmd.Flags().StringVarP(&createLmFlags.serviceType, "service-type", "s", "ClusterIP", "service type: ClusterIP, NodePort, LoadBalancer, ExternalName")
	createLmCmd.Flags().StringVarP(&createLmFlags.cpu, "cpu", "c", "4", "cpu")
	createLmCmd.Flags().BoolVar(&createLmFlags.export, "export", false, "export manifests instead of installing")
	createLmCmd.Flags().BoolVar(&createLmFlags.wait, "wait", false, "wait for the resources to be reconciled")

	rootCmd.AddCommand(createLmCmd)
}

func createLmCmdRun(cmd *cobra.Command, args []string) error {
	lmName := args[0]
	lmtemplate := `---
apiVersion: ai.contrib.fluxcd.io/v1alpha1
kind: LanguageModel
metadata:
  name: {{ .LMName }}
  namespace: {{ .LMNamespace }}
spec:
  sourceRef:
    kind: OCIRepository
    name: {{ .Model }}
    namespace: {{ .ModelNamespace }}
  interval: 10m
  timeout: 2m
  prune: true
  engine:
    serviceType: {{ .ServiceType }}
    replicas: 1
    resources:
      requests:
        cpu: "{{ .CPU }}"
`
	tpl, err := template.New("create-lm").Parse(lmtemplate)
	if err != nil {
		return err
	}
	data := struct {
		LMName         string
		LMNamespace    string
		Model          string
		ModelNamespace string
		ServiceType    string
		CPU            string
	}{
		LMName:         lmName,
		LMNamespace:    *kubeconfigArgs.Namespace,
		Model:          createLmFlags.model,
		ModelNamespace: createLmFlags.modelNamespace,
		ServiceType:    createLmFlags.serviceType,
		CPU:            createLmFlags.cpu,
	}

	var buffer bytes.Buffer
	if err := tpl.Execute(&buffer, data); err != nil {
		return err
	}

	if createLmFlags.export {
		fmt.Print(buffer.String())
		return nil
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), rootArgs.timeout)
	defer cancelFn()

	applyOutput, err := utils.Apply(ctx, kubeconfigArgs, kubeclientOptions, buffer.Bytes(), func(e ssa.ChangeSetEntry) (wait bool) {
		return true
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, applyOutput)

	return nil
}
