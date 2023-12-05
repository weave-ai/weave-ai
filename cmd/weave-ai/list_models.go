package main

import (
	"context"
	"fmt"
	"github.com/dustin/go-humanize"
	meta2 "github.com/fluxcd/pkg/apis/meta"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/spf13/cobra"
	"github.com/weave-ai/weave-ai/pkg/utils"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"text/tabwriter"
)

var listModelsCmd = &cobra.Command{
	Use:     "list-models",
	Aliases: []string{"list-model", "models"},
	Short:   "List all OCI Language Model resources",
	RunE:    listModelsCmdRun,
}

var listModelsFlags struct {
	all bool
}

func init() {
	listModelsCmd.Flags().BoolVarP(&listModelsFlags.all, "all", "A", false, "Show models from all namespaces")
	rootCmd.AddCommand(listModelsCmd)
}

func listModelsCmdRun(cmd *cobra.Command, args []string) error {
	cli, err := utils.KubeClient(kubeconfigArgs, kubeclientOptions)
	if err != nil {
		return err
	}

	namespace := *kubeconfigArgs.Namespace
	if listModelsFlags.all {
		namespace = ""
	}

	models := &sourcev1b2.OCIRepositoryList{}
	ctx, _ := context.WithTimeout(context.Background(), rootArgs.timeout)
	if err := cli.List(ctx, models, client.InNamespace(namespace), client.MatchingLabels{
		"ai.contrib.fluxcd.io/artifact-kind": "language-model",
	}); err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "NAME\tVERSION\tFAMILY\tSTATUS\tCREATED\n")
	for _, model := range models.Items {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			model.Namespace+"/"+model.Name,
			model.Spec.Reference.Tag,
			getFamily(model),
			getStatus(model),
			humanize.Time(model.CreationTimestamp.Time),
		)
	}
	w.Flush()

	return nil
}

func getFamily(model sourcev1b2.OCIRepository) any {
	if model.Status.Artifact == nil {
		return ""
	}

	return model.Status.Artifact.Metadata["ai.contrib.fluxcd.io/family"]
}

func getStatus(model sourcev1b2.OCIRepository) string {
	if model.Spec.Suspend == true {
		return "INACTIVE"
	}

	if model.Status.Artifact == nil {
		return "INACTIVE"
	}

	cond := meta.FindStatusCondition(model.Status.Conditions, meta2.ReadyCondition)
	if cond == nil {
		return "INACTIVE"
	}

	if cond.Status == metav1.ConditionFalse {
		return "NOT READY"
	}

	if model.Status.Artifact != nil && model.Status.Artifact.URL != "" {
		return "* ACTIVE"
	}

	return "UNKNOWN"
}
