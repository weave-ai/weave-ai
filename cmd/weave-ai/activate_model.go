package main

import (
	"context"
	"strings"
	"time"

	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/spf13/cobra"
	"github.com/weave-ai/weave-ai/pkg/utils"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var activateModelCmd = &cobra.Command{
	Use:   "activate-model",
	Args:  cobra.ExactArgs(1),
	Short: "Activate a model",
	Long: `
# Activate the zephyr-7b-beta model from the weave-ai namespace
weave-ai activate-model zephyr-7b-beta

# Activate the zephyr-7b-beta model from the weave-ai namespace
weave-ai activate-model weave-ai/zephyr-7b-beta

# Activate the zephyr-7b-beta model and wait for it to be activated
weave-ai activate-model --wait zephyr-7b-beta
`,
	RunE: activateModelCmdRun,
}

var activateModelFlags struct {
	modelName      string
	modelNamespace string
	wait           bool
}

func init() {
	activateModelCmd.Flags().BoolVarP(&activateModelFlags.wait, "wait", "w", true, "Wait for the model to be activated")
	rootCmd.AddCommand(activateModelCmd)
}

func activateModelCmdRun(cmd *cobra.Command, args []string) error {
	modelName := args[0]
	// if model name contains / split it into model namespace and model name
	if strings.Contains(modelName, "/") {
		split := strings.SplitN(modelName, "/", 2)
		activateModelFlags.modelNamespace = split[0]
		activateModelFlags.modelName = split[1]
	} else {
		activateModelFlags.modelName = modelName
		activateModelFlags.modelNamespace = defaultNamespace
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), rootArgs.timeout)
	defer cancelFn()

	client, err := utils.KubeClient(kubeconfigArgs, kubeclientOptions)
	if err != nil {
		return err
	}

	if err := activateModel(ctx,
		client,
		activateModelFlags.modelNamespace,
		activateModelFlags.modelName,
		activateModelFlags.wait); err != nil {
		return err
	}

	return nil
}

func activateModel(ctx context.Context, client runtimeclient.Client, namespace string, name string, waitFlag bool) error {
	logger.Actionf("checking if model %s/%s exists and is active", namespace, name)
	// check the model exists
	model := &sourcev1b2.OCIRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OCIRepository",
			APIVersion: "source.toolkit.fluxcd.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	// check the model is ready
	if err := client.Get(ctx, runtimeclient.ObjectKeyFromObject(model), model); err != nil {
		return err
	}

	// try to activate the model if it's not active
	if model.Spec.Suspend == true {
		logger.Actionf("activate model %s/%s", namespace, name)
		model.Spec.Suspend = false
		if err := client.Update(ctx, model); err != nil {
			return err
		}
	} else {
		logger.Successf("model %s/%s is already active", namespace, name)
		return nil
	}

	if waitFlag {
		logger.Waitingf("waiting for model %s/%s to be active", namespace, name)

		waitCtx, waitCancel := context.WithCancel(ctx)
		wait.UntilWithContext(waitCtx, func(ctx context.Context) {
			if err := client.Get(ctx, runtimeclient.ObjectKeyFromObject(model), model); err != nil {
				return
			}
			if model.Status.Artifact == nil {
				return
			}
			if model.Status.Artifact.URL == "" {
				return
			}
			cond := apimeta.FindStatusCondition(model.Status.Conditions, fluxmeta.ReadyCondition)
			if cond == nil {
				return
			}
			if cond.Status != metav1.ConditionTrue {
				return
			}
			if model.Status.Artifact.URL != "" {
				waitCancel()
			}
		}, 2*time.Second)
	}

	// TODO if it's not ready after 5 minutes, return an error

	return nil
}
