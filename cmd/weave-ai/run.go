package main

import (
	"context"
	"fmt"
	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"github.com/weave-ai/weave-ai/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/wait"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"

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
# Deploy and run an LLM, e.g. zephyr-7b-beta, in the default namespace.
# The default model namespace is weave-ai.
weave-ai run zephyr-7b-beta

# Deploy and run an LLM, e.g. zephyr-7b-beta, in the default namespace as my-llm.
# The default model namespace is weave-ai.
weave-ai run --name=my-llm zephyr-7b-beta

# Deploy and run an LLM, e.g. zephyr-7b-beta from the weave-ai model namespace, in the default namespace.
weave-ai run weave-ai/zephyr-7b-beta

# Deploy and run an LLM, e.g. zephyr-7b-beta, in the default namespace and publish it as a LoadBalancer service.
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
	ui             bool // start the UI
	local          bool // run the LLM locally
}

func init() {
	runCmd.Flags().StringVar(&runFlags.name, "name", "", "name of the LLM")
	runCmd.Flags().BoolVarP(&runFlags.publish, "publish", "p", false, "publish the LLM, which means it will be exposed as a LoadBalancer service")
	runCmd.Flags().StringVarP(&runFlags.cpu, "cpu", "c", "4", "cpu")
	runCmd.Flags().BoolVarP(&runFlags.detach, "detach", "d", false, "detach from the process e.g. not follow the logs")
	runCmd.Flags().BoolVar(&runFlags.ui, "ui", false, "start the Weave Chat UI along side the LLM")
	// runCmd.Flags().BoolVar(&runFlags.local, "local", false, "run the LLM locally")

	// TODO use the default namespace from context
	runCmd.Flags().StringVarP(&runFlags.namespace, "namespace", "n", "default", "namespace")
	rootCmd.AddCommand(runCmd)
}

func runCmdRun(cmd *cobra.Command, args []string) error {
	if runFlags.local {
		return runCmdRunLocal(cmd, args)
	}
	return runCmdRun0(cmd, args)
}

func runCmdRunLocal(cmd *cobra.Command, args []string) error {
	/*
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

		// 1. check ~/.weave-ai/bin
		// 2. download https://github.com/Mozilla-Ocho/llamafile/releases/download/0.4/llamafile-0.4
		// 3. rename to llamafile
		// 4. chmod +x llamafile
		// 5. move to ~/.weave-ai/bin
		// 6. run llamafile run
	*/
	return nil
}

func runCmdRun0(cmd *cobra.Command, args []string) error {
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
			Labels: map[string]string{
				"ai.contrib.fluxcd.io/model-namespace": runFlags.modelNamespace,
				"ai.contrib.fluxcd.io/model":           runFlags.modelName,
			},
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

	if err := activateModel(ctx, client, runFlags.modelNamespace, runFlags.modelName, true); err != nil {
		return err
	}

	logger.Actionf("creating new LLM instance %s/%s", runFlags.namespace, lmName)
	if err := client.Create(ctx, lm); err != nil {
		return err
	}

	logger.Waitingf("waiting for %s/%s to be ready", runFlags.namespace, lmName)
	waitCtx, waitCancel := context.WithCancel(ctx)
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

	var (
		ui    *appsv1.Deployment
		uiSvc *corev1.Service
	)

	if runFlags.ui {
		uiAppName := lmName + "-chat-app"
		clusterDomain := rootArgs.clusterDomain
		ownerRefs := []metav1.OwnerReference{
			{
				APIVersion:         "ai.contrib.fluxcd.io/v1alpha1",
				Kind:               "LanguageModel",
				Name:               lm.Name,
				UID:                lm.UID,
				BlockOwnerDeletion: &[]bool{true}[0],
				Controller:         &[]bool{true}[0],
			},
		}

		labels := map[string]string{"app": uiAppName}
		ui = &appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            uiAppName,
				Namespace:       runFlags.namespace,
				Labels:          labels,
				OwnerReferences: ownerRefs,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &[]int32{1}[0],
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: labels,
					},
					Spec: corev1.PodSpec{
						SecurityContext: &corev1.PodSecurityContext{
							RunAsUser:    &[]int64{65532}[0],
							RunAsNonRoot: &[]bool{true}[0],
						},
						Containers: []corev1.Container{
							{
								Name:  "chat-app",
								Image: ImageChatInfo,
								Env: []corev1.EnvVar{
									{
										Name:  "LLM_API_HOST",
										Value: svc.Name + "." + runFlags.namespace + ".svc." + clusterDomain + ":8000",
									},
								},
								SecurityContext: &corev1.SecurityContext{
									Privileged:   &[]bool{false}[0],
									RunAsNonRoot: &[]bool{true}[0],
									RunAsUser:    &[]int64{65532}[0],
									Capabilities: &corev1.Capabilities{
										Drop: []corev1.Capability{
											"ALL",
										},
									},
								},
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: 8501,
										Name:          "http",
										Protocol:      corev1.ProtocolTCP,
									},
								},
							},
						},
					},
				},
			},
		}
		if err := client.Create(ctx, ui); err != nil {
			return err
		}

		uiSvc = &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            uiAppName,
				Namespace:       runFlags.namespace,
				Labels:          labels,
				OwnerReferences: ownerRefs,
			},
			Spec: corev1.ServiceSpec{
				Selector: labels,
				Type:     corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{
					{
						Name:       "http",
						Port:       8501,
						TargetPort: intstr.FromInt32(8501),
					},
				},
			},
		}

		if err := client.Create(ctx, uiSvc); err != nil {
			return err
		}

		// wait is good for the UI to be ready
		logger.Waitingf("waiting for %s/%s to be ready", runFlags.namespace, uiAppName)
		waitCtx, waitCancel := context.WithCancel(ctx)
		wait.UntilWithContext(waitCtx, func(ctx context.Context) {
			if err := client.Get(ctx, runtimeclient.ObjectKeyFromObject(ui), ui); err != nil {
				return
			}
			var cond *appsv1.DeploymentCondition
			for _, condition := range ui.Status.Conditions {
				if condition.Type == appsv1.DeploymentAvailable {
					cond = &condition
					break
				}
			}
			if cond == nil {
				return
			}
			if cond.Status != corev1.ConditionTrue {
				return
			}
			waitCancel()
		}, 2*time.Second)

	}

	if !runFlags.detach {
		pods := &corev1.PodList{}

		// engine pods
		matchingLabels := runtimeclient.MatchingLabels{"app": lm.Name}
		podLogOpts := corev1.PodLogOptions{
			Container: "engine",
			Follow:    true,
		}

		// if UI is enabled, wait for the UI pod to be ready
		if runFlags.ui {
			matchingLabels["app"] = lm.Name + "-chat-app"
			podLogOpts.Container = "chat-app"
		}

		if err := client.List(ctx, pods,
			runtimeclient.InNamespace(runFlags.namespace),
			matchingLabels,
		); err != nil {
			return err
		}

		podName := pods.Items[0].Name
		config, err := utils.KubeConfig(kubeconfigArgs, kubeclientOptions)
		if err != nil {
			return err
		}

		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			return err
		}
		req := clientSet.CoreV1().Pods(runFlags.namespace).GetLogs(podName, &podLogOpts)

		var podLogs io.ReadCloser
		i := 0
		const MaxRetry = 10
		for {
			var err error
			podLogs, err = req.Stream(context.Background())
			if err != nil {
				time.Sleep(1 * time.Second)
				i = i + 1
				if i > MaxRetry {
					return err
				}
			}
			break
		}
		if podLogs == nil {
			return fmt.Errorf("could not get logs for pod %s/%s", runFlags.namespace, podName)
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
	} else {
		// if detached, shows kubectl port-forward commands
		logger.Successf("to connect to your LLM:\n  kubectl port-forward -n %s svc/%s 8000:8000", svc.Namespace, svc.Name)
		if runFlags.ui {
			logger.Successf("to connect to the UI:\n  kubectl port-forward -n %s svc/%s 8501:8501", uiSvc.Namespace, uiSvc.Name)
		}
	}

	return nil
}
