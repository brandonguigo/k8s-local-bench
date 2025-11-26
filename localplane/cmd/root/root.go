/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package rootCmd

import (
	"errors"
	clusterCmd "localplane/cmd/cluster"
	"localplane/config"
	"localplane/utils/viperutils"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var CfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "localplane",
	Short: "a cli tool to run a kubernetes cluster locally for development and testing",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig(cmd)
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.PersistentFlags().StringP("directory", "d", ".", "Directory where configurations and data are stored")
	viperutils.MapFlagToEnv(rootCmd, "directory", "LOCALPLANE_DIRECTORY", "directory")
	rootCmd.PersistentFlags().StringVarP(&CfgFile, "config", "c", "", "config file (default is /.localplane.yaml)")

	rootCmd.AddCommand(clusterCmd.NewCommand())
}

func initializeConfig(cmd *cobra.Command) error {
	// 1. Set up Viper to use environment variables.
	viper.SetEnvPrefix("LOCALPLANE")

	// Allow for nested keys in environment variables (e.g. `LOCALPLANE_DATABASE_HOST`)
	// Replace dots and dashes with underscores so keys like `selfhost.directory`
	// map to env vars like `LOCALPLANE_SELFHOST_DIRECTORY`.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	// 2. Bind Cobra flags to Viper early so flags (and their defaults) and env vars
	// are available when we need to determine paths (e.g. the --directory flag
	// is used to locate the config file). Bind both the command-local flags and
	// the root persistent flags so flags passed to subcommands are recognized.
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}
	if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
		return err
	}
	// Also ensure root-level persistent flags are bound (covers some cobra versions
	// where cmd.PersistentFlags() may not include parent persistent flags).
	if root := cmd.Root(); root != nil {
		if err := viper.BindPFlags(root.PersistentFlags()); err != nil {
			return err
		}
	}

	// Prefer explicit config file override via --config flag.
	if CfgFile != "" {
		viper.SetConfigFile(CfgFile)
	} else {
		// Prefer a hidden file named .localplane-config.yaml if present in the directory.
		viper.AddConfigPath("$HOME")
		viper.SetConfigName(".localplane")
		viper.SetConfigType("yaml")
	}

	// 3. Read the configuration file.
	// If a config file is found, read it in. We use a robust error check
	// to ignore "file not found" errors, but panic on any other error.
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist.
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return err
		}
	}

	// 5. Unmarshal the configuration into the CliConfig struct.
	err := viper.Unmarshal(&config.CliConfig)
	if err != nil {
		return err
	}

	debug := os.Getenv("LOG_LEVEL")

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug == "debug" || config.CliConfig.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Debug().Interface("config", config.CliConfig).Msg("Configuration initialized successfully")

	return nil
}
