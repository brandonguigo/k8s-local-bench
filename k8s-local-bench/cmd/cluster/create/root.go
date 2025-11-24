package create

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s-local-bench/config"
	kindsvc "k8s-local-bench/utils/kind"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewCommand creates the cluster command
func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create",
		Short: "create a local k8s cluster",
		Run:   createCluster,
	}
	// flags
	cmd.Flags().StringP("kind-config", "k", "", "path to kind config file (searched in current directory if unspecified)")
	cmd.Flags().BoolP("yes", "y", false, "don't ask for confirmation; assume yes")
	cmd.Flags().Bool("start-lb", true, "start local load balancer (cloud-provider-kind)")
	cmd.Flags().Bool("lb-foreground", false, "run load balancer in foreground (blocking)")
	// add subcommands here
	return cmd
}

func createCluster(cmd *cobra.Command, args []string) {
	log.Info().Msg("Creating local k8s cluster...")
	// honor CLI debug config if set
	if config.CliConfig.Debug {
		log.Debug().Bool("debug", true).Msg("debug enabled")
	}

	// check for kind config file
	kindCfg, _ := cmd.Flags().GetString("kind-config")
	if kindCfg == "" {
		kindCfg = findKindConfig()
		if kindCfg == "" {
			log.Info().Msg("no kind config file found in current directory; proceeding without one")
		} else {
			log.Info().Str("path", kindCfg).Msg("found kind config file in current directory")
		}
	} else {
		if _, err := os.Stat(kindCfg); err != nil {
			log.Warn().Err(err).Str("path", kindCfg).Msg("provided kind config not found; ignoring")
			kindCfg = ""
		} else {
			log.Info().Str("path", kindCfg).Msg("using kind config from flag")
		}
	}

	// ask for confirmation unless user passed --yes
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Printf("Proceed to create kind cluster '%s'? (y/N): ", "local-bench")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if !(strings.EqualFold(input, "y") || strings.EqualFold(input, "yes")) {
			log.Info().Msg("aborting cluster creation")
			return
		}
	}

	// create a kind cluster (name is currently fixed)
	clusterName := "local-bench"
	if err := kindsvc.Create(clusterName, kindCfg); err != nil {
		log.Error().Err(err).Msg("failed creating kind cluster")
		return
	}
	log.Info().Str("name", clusterName).Msg("kind cluster creation invoked")
	// start load balancer if requested (defaults: start and run in background)
	startLB, _ := cmd.Flags().GetBool("start-lb")
	lbFg, _ := cmd.Flags().GetBool("lb-foreground")
	if startLB {
		if !lbFg {
			// background
			if err := kindsvc.StartLoadBalancer(clusterName, true); err != nil {
				log.Error().Err(err).Msg("failed to start load balancer in background")
			} else {
				log.Info().Msg("load balancer started in background")
			}
		} else {
			// foreground: run and wait (this will block until the process exits)
			if err := kindsvc.StartLoadBalancer(clusterName, false); err != nil {
				log.Error().Err(err).Msg("failed to run load balancer (foreground)")
			} else {
				log.Info().Msg("load balancer run completed")
			}
		}
	}

	// TODO: make sure the cluster is up and running via kubectl commands
}

// findKindConfig searches the current working directory for common kind config filenames.
// Returns the first match (absolute path) or empty string if none found.
func findKindConfig() string {
	base := config.CliConfig.Directory
	var err error
	if base == "" {
		base, err = os.Getwd()
		if err != nil {
			return ""
		}
	}
	candidates := []string{"kind-config.yaml", "kind-config.yml",}
	for _, name := range candidates {
		p := filepath.Join(base, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// try broader glob for files starting with "kind"
	matches, _ := filepath.Glob(filepath.Join(base, "kind*.y*ml"))
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}
