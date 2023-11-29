package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fluxcd/pkg/ssa"
	"io"
	"os"
	"text/template"
	"time"

	"github.com/fluxcd/flux2/v2/pkg/status"
	"github.com/spf13/cobra"
	"github.com/weave-ai/weave-ai/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the Weave AI controllers",
	Long: fmt.Sprintf(`
# Install the Weave AI controllers
weave-ai install
`),
	RunE: installCmdRun,
}

var installFlags struct {
	version string
	// dev       bool
	// anonymous bool
	export bool
	// mode      string
	withModelCatalog bool
}

func init() {
	installCmd.Flags().StringVarP(&installFlags.version, "version", "v", Version, "version of Weave AI to install")
	installCmd.Flags().BoolVar(&installFlags.export, "export", false, "export manifests instead of installing")
	installCmd.Flags().BoolVar(&installFlags.withModelCatalog, "with-model-catalog", true, "install the model catalog")

	rootCmd.AddCommand(installCmd)
}

func installCmdRun(cmd *cobra.Command, args []string) error {
	if installFlags.export {
		logger.stderr = io.Discard
	}

	if err := installControllers(installFlags.export, installFlags.version, installFlags.withModelCatalog); err != nil {
		return err
	}

	if !installFlags.export {
		if err := verifyTheInstallation(); err != nil {
			return err
		}
	}

	return nil
}

func buildComponentObjectRefs(namespace string, components ...string) ([]object.ObjMetadata, error) {
	var objRefs []object.ObjMetadata
	for _, deployment := range components {
		objRefs = append(objRefs, object.ObjMetadata{
			Namespace: namespace,
			Name:      deployment,
			GroupKind: schema.GroupKind{Group: "apps", Kind: "Deployment"},
		})
	}
	return objRefs, nil
}

func verifyTheInstallation() error {
	logger.Waitingf("verifying installation")

	kubeConfig, err := utils.KubeConfig(kubeconfigArgs, kubeclientOptions)
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	statusChecker, err := status.NewStatusChecker(kubeConfig, 5*time.Second, rootArgs.timeout, logger)
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	objectRefs, err := buildComponentObjectRefs(
		*kubeconfigArgs.Namespace,
		"lm-controller",
	)
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	if err := statusChecker.Assess(objectRefs...); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	logger.Successf("install finished")
	return nil
}

func installControllers(export bool, version string, withModelCatalog bool) error {
	logger.Generatef("generating manifests")

	var tpl bytes.Buffer
	t, err := template.New("template").Parse(installTemplate)
	if err != nil {
		return err
	}

	if err := t.Execute(&tpl, struct {
		ModelCatalog bool
		Version      string
	}{
		ModelCatalog: withModelCatalog,
		Version:      version,
	}); err != nil {
		return err
	}

	// Use Kustomize (krusty) to build the kustomization
	fSys := filesys.MakeFsInMemory()
	kustomizationPath := "/app/kustomization.yaml"
	fSys.WriteFile(kustomizationPath, tpl.Bytes())

	namespacePath := "/app/namespace.yaml"
	fSys.WriteFile(namespacePath, []byte(fmt.Sprintf(namespaceTemplate, *kubeconfigArgs.Namespace)))

	/*
		clusterPath := "/app/cluster.yaml"
		if installMode == TenantMode {
			fSys.WriteFile(clusterPath, []byte(
				fmt.Sprintf(defaultClusterSecretTemplate,
					kubeconfigArgs.Namespace,
					kubeconfigArgs.Namespace,
					kubeconfigArgs.Namespace,
					kubeconfigArgs.Namespace,
					kubeconfigArgs.Namespace,
				)))
		} else {
			fSys.WriteFile(clusterPath, []byte("# empty"))
		}
	*/

	opts := krusty.MakeDefaultOptions()
	opts.Reorder = krusty.ReorderOptionLegacy
	k := krusty.MakeKustomizer(opts)

	m, err := k.Run(fSys, "/app")
	if err != nil {
		return err
	}

	yamlOutput, err := m.AsYaml()
	if err != nil {
		return err
	}
	logger.Successf("manifests build completed")

	if export {
		fmt.Println(string(yamlOutput))
		return nil
	}

	// install everything
	logger.Actionf("installing components in %s namespace", *kubeconfigArgs.Namespace)

	ctx, cancelFn := context.WithTimeout(context.Background(), rootArgs.timeout)
	defer cancelFn()

	applyOutput, err := utils.Apply(ctx, kubeconfigArgs, kubeclientOptions, yamlOutput, func(e ssa.ChangeSetEntry) (wait bool) {
		wait = true
		if e.ObjMetadata.GroupKind.Kind == "OCIRepository" {
			wait = false
		}
		return
	})
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}
	fmt.Fprintln(os.Stderr, applyOutput)

	return nil
}
