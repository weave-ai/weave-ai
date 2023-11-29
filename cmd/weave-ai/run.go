package main

import (
	"context"
	"fmt"
	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	"k8s.io/client-go/kubernetes"

	"github.com/weave-ai/weave-ai/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/wait"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"

	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/spf13/cobra"
	aiv1a1 "github.com/weave-ai/lm-controller/api/v1alpha1"
	namesgenerator "github.com/weave-ai/weave-ai/pkg/namegenerator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Args:  cobra.ExactArgs(1),
	Short: "Deploy and run an LLM",
	Long: `
# Deploy and run an LLM, e.g. zephyr-7b-beta, in the default namespace
weave-ai run zephyr-7b-beta

# Deploy and run an LLM, e.g. zephyr-7b-beta from the weave-ai model namespace, in the default namespace
weave-ai run weave-ai/zephyr-7b-beta

# Deploy and run an LLM, e.g. zephyr-7b-beta, in the default namespace and publish it as a LoadBalancer service
weave-ai run -p -d weave-ai/zephyr-7b-beta
`,
	RunE: runCmdRun,
}

var runFlags struct {
	namespace      string
	name           string // name of the LLM
	publish        bool   // publish the LLM, which means it will be exposed as a LoadBalancer service
	cpu            string
	modelName      string
	modelNamespace string
	detach         bool // detach from the process e.g. not follow the logs
}

func init() {
	runCmd.Flags().StringVar(&runFlags.name, "name", "", "name of the LLM")
	runCmd.Flags().BoolVarP(&runFlags.publish, "publish", "p", false, "publish the LLM, which means it will be exposed as a LoadBalancer service")
	runCmd.Flags().StringVarP(&runFlags.cpu, "cpu", "c", "4", "cpu")
	runCmd.Flags().BoolVarP(&runFlags.detach, "detach", "d", false, "detach from the process e.g. not follow the logs")
	// TODO use the default namespace from context
	runCmd.Flags().StringVarP(&runFlags.namespace, "namespace", "n", "default", "namespace")
	rootCmd.AddCommand(runCmd)
}

func runCmdRun(cmd *cobra.Command, args []string) error {
	modelName := args[0]
	// if model name contains / split it into model namespace and model name
	if strings.Contains(modelName, "/") {
		split := strings.SplitN(modelName, "/", 2)
		runFlags.modelNamespace = split[0]
		runFlags.modelName = split[1]
	} else {
		runFlags.modelName = modelName
		runFlags.modelNamespace = defaultNamespace
	}

	lmName := runFlags.name
	if lmName == "" {
		// random name using the docker name lib
		lmName = namesgenerator.GetRandomName(0)
	}

	serviceType := "ClusterIP"
	if runFlags.publish {
		serviceType = "LoadBalancer"
	}

	lm := &aiv1a1.LanguageModel{
		TypeMeta: metav1.TypeMeta{
			Kind:       "LanguageModel",
			APIVersion: "ai.contrib.fluxcd.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      lmName,
			Namespace: runFlags.namespace,
		},
		Spec: aiv1a1.LanguageModelSpec{
			SourceRef: aiv1a1.CrossNamespaceSourceReference{
				Kind:      "OCIRepository",
				Name:      runFlags.modelName,
				Namespace: runFlags.modelNamespace,
			},
			Interval:      metav1.Duration{Duration: 2 * time.Minute},
			RetryInterval: metav1.Duration{Duration: 30 * time.Second},
			Timeout:       &metav1.Duration{Duration: 2 * time.Minute},
			Prune:         true,
			Engine: aiv1a1.EngineSpec{
				ServiceType: corev1.ServiceType(serviceType),
				Replicas:    &[]int32{1}[0],
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse(runFlags.cpu),
					},
				},
			},
		},
	}

	if createLmFlags.export {
		// TODO export manifests instead of installing
		// fmt.Print(buffer.String())
		return nil
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), rootArgs.timeout)
	defer cancelFn()

	client, err := utils.KubeClient(kubeconfigArgs, kubeclientOptions)
	if err != nil {
		return err
	}

	logger.Actionf("checking if model %s/%s exists and is active", runFlags.modelNamespace, runFlags.modelName)
	// check the model exists
	model := &sourcev1b2.OCIRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OCIRepository",
			APIVersion: "source.toolkit.fluxcd.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      runFlags.modelName,
			Namespace: runFlags.modelNamespace,
		},
	}

	// check the model is ready
	if err := client.Get(ctx, runtimeclient.ObjectKeyFromObject(model), model); err != nil {
		return err
	}
	// try to activate the model if it's not active
	if model.Spec.Suspend == true {
		logger.Actionf("activate model %s/%s", runFlags.modelNamespace, runFlags.modelName)
		model.Spec.Suspend = false
		if err := client.Update(ctx, model); err != nil {
			return err
		}

		logger.Waitingf("waiting for model %s/%s to be active", runFlags.modelNamespace, runFlags.modelName)
	}

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
	// TODO if it's not ready after 5 minutes, return an error

	logger.Actionf("creating new LLM instance %s/%s", runFlags.namespace, lmName)
	if err := client.Create(ctx, lm); err != nil {
		return err
	}

	logger.Waitingf("waiting for %s/%s to be ready", runFlags.namespace, lmName)
	waitCtx, waitCancel = context.WithCancel(ctx)
	wait.UntilWithContext(waitCtx, func(ctx context.Context) {
		if err := client.Get(ctx, runtimeclient.ObjectKeyFromObject(lm), lm); err != nil {
			return
		}
		cond := apimeta.FindStatusCondition(lm.Status.Conditions, fluxmeta.ReadyCondition)
		if cond == nil {
			return
		}
		if cond.Status != metav1.ConditionTrue {
			return
		}
		waitCancel()
	}, 2*time.Second)

	svc := &corev1.Service{}
	if err := client.Get(ctx, runtimeclient.ObjectKeyFromObject(lm), svc); err != nil {
		return err
	}

	if runFlags.publish {
		logger.Waitingf("waiting for language model %s/%s to be published", runFlags.namespace, lmName)
		waitCtx, waitCancel := context.WithCancel(ctx)
		wait.UntilWithContext(waitCtx, func(ctx context.Context) {
			if err := client.Get(ctx, runtimeclient.ObjectKeyFromObject(svc), svc); err != nil {
				return
			}
			if len(svc.Status.LoadBalancer.Ingress) == 0 {
				return
			}
			if svc.Status.LoadBalancer.Ingress[0].IP == "" {
				return
			}
			waitCancel()
		}, 2*time.Second)
		ip := svc.Status.LoadBalancer.Ingress[0].IP
		logger.Successf("your LLM is ready at http://%s:8000", ip)
	}

	if !runFlags.detach {
		pod := &corev1.PodList{}
		if err := client.List(ctx, pod,
			runtimeclient.InNamespace(runFlags.namespace),
			runtimeclient.MatchingLabels{"app": lm.Name}); err != nil {
			return err
		}

		podName := pod.Items[0].Name
		config, err := utils.KubeConfig(kubeconfigArgs, kubeclientOptions)
		if err != nil {
			return err
		}

		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			return err
		}
		podLogOpts := corev1.PodLogOptions{
			Container: "engine",
			Follow:    true,
		}
		req := clientSet.CoreV1().Pods(runFlags.namespace).GetLogs(podName, &podLogOpts)
		podLogs, err := req.Stream(context.Background())
		if err != nil {
			return err
		}
		defer podLogs.Close()

		// Stream the pod logs
		buf := make([]byte, 2000)
		for {
			w, err := podLogs.Read(buf)
			if w > 0 {
				fmt.Print(string(buf[:w]))
			}
			if err != nil {
				break
			}
		}
	}

	return nil
}
