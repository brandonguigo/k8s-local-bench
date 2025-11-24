package create

import (
	"os"
	"path/filepath"

	"k8s-local-bench/config"

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
	// flag to allow specifying a kind config file; if empty we'll look in CWD
	cmd.Flags().StringP("kind-config", "k", "", "path to kind config file (searched in current directory if unspecified)")
	// add subcommands here
	return cmd
}

func createCluster(cmd *cobra.Command, args []string) {
	log.Info().Msg("Creating local k8s cluster...")
	// honor CLI debug config if set
	if config.CliConfig.Debug {
		log.Debug().Bool("debug", true).Msg("debug enabled")
	}

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

	// TODO: create kind cluster based on confif if exists

	// TODO: run the cloud-provider-kind command to have loadbalancer support in a separate thread

	// TODO: make sure the cluster is up and running via kubectl commands
}

// findKindConfig searches the current working directory for common kind config filenames.
// Returns the first match (absolute path) or empty string if none found.
func findKindConfig() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	candidates := []string{"kind-config.yaml", "kind-config.yml", "kind.yaml", "kind.yml"}
	for _, name := range candidates {
		p := filepath.Join(cwd, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// try broader glob for files starting with "kind"
	matches, _ := filepath.Glob(filepath.Join(cwd, "kind*.y*ml"))
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}
