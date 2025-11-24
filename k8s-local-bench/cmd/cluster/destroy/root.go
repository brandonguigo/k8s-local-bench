package destroy

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
		Use:   "destroy",
		Short: "destroy a local k8s cluster",
		Run:   createCluster,
	}
	// add subcommands here
	return cmd
}

func createCluster(cmd *cobra.Command, args []string) {
	log.Info().Msg("Deleting local k8s cluster...")
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

	// TODO: delete a kind cluster (name is currently fixed)

	// TODO: stop the cloud-provider-kind command

	// TODO: make sure the cluster is stopped/deleted
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
