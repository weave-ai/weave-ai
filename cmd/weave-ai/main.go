package main

import (
	"fmt"
	"github.com/go-logr/logr"
	"log"
	"os"
	"time"

	runclient "github.com/fluxcd/pkg/runtime/client"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

var rootCmd = &cobra.Command{
	Use:           "weave-ai",
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	Short:         "Weave AI CLI",
	Long: `Weave AI CLI - the command line interface for Weave AI.

# Install Weave AI.
weave-ai install
`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			return fmt.Errorf("error getting namespace: %w", err)
		}

		if e := validation.IsDNS1123Label(ns); len(e) > 0 {
			return fmt.Errorf("namespace must be a valid DNS label: %q", ns)
		}

		return nil
	},
}

var logger = stderrLogger{stderr: os.Stderr}

type rootFlags struct {
	timeout      time.Duration
	verbose      bool
	pollInterval time.Duration
}

const defaultNamespace = "weave-ai"

var (
	rootArgs          = newRootFlags()
	kubeconfigArgs    = genericclioptions.NewConfigFlags(false)
	kubeclientOptions = new(runclient.Options)
)

func init() {
	rootCmd.PersistentFlags().DurationVar(&rootArgs.timeout, "timeout", 2*time.Minute, "timeout for this operation")
	rootCmd.PersistentFlags().BoolVar(&rootArgs.verbose, "verbose", false, "print generated objects")

	configureDefaultNamespace()
	kubeconfigArgs.APIServer = nil // prevent AddFlags from configuring --server flag
	kubeconfigArgs.Timeout = nil   // prevent AddFlags from configuring --request-timeout flag, we have --timeout instead
	kubeconfigArgs.AddFlags(rootCmd.PersistentFlags())

	// Since some subcommands use the `-s` flag as a short version for `--silent`, we manually configure the server flag
	// without the `-s` short version. While we're no longer on par with kubectl's flags, we maintain backwards compatibility
	// on the CLI interface.
	apiServer := ""
	kubeconfigArgs.APIServer = &apiServer
	rootCmd.PersistentFlags().StringVar(kubeconfigArgs.APIServer, "server", *kubeconfigArgs.APIServer, "The address and port of the Kubernetes API server")

	kubeclientOptions.BindFlags(rootCmd.PersistentFlags())

	rootCmd.DisableAutoGenTag = true
	rootCmd.SetOut(os.Stdout)
}

func newRootFlags() rootFlags {
	rf := rootFlags{
		pollInterval: 2 * time.Second,
	}
	return rf
}

func configureDefaultNamespace() {
	*kubeconfigArgs.Namespace = defaultNamespace
	fromEnv := os.Getenv("WEAVE_AI_NAMESPACE")
	if fromEnv != "" {
		// namespace must be a valid DNS label. Assess against validation
		// used upstream, and ignore invalid values as environment vars
		// may not be actively provided by end-user.
		if e := validation.IsDNS1123Label(fromEnv); len(e) > 0 {
			logger.Failuref(" ignoring invalid WEAVE_AI_NAMESPACE: %v", fromEnv)
			return
		}

		kubeconfigArgs.Namespace = &fromEnv
	}
}

func main() {
	log.SetFlags(0)

	// This is required because controller-runtime expects its consumers to
	// set a logger through log.SetLogger within 30 seconds of the program's
	// initalization. If not set, the entire debug stack is printed as an
	// error, see: https://github.com/kubernetes-sigs/controller-runtime/blob/ed8be90/pkg/log/log.go#L59
	// Since we have our own logging and don't care about controller-runtime's
	// logger, we configure it's logger to do nothing.
	ctrllog.SetLogger(logr.New(ctrllog.NullLogSink{}))

	if err := rootCmd.Execute(); err != nil {
		logger.Failuref("%v", err)
		os.Exit(1)
	}
}
