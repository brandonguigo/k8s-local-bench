package viperutils

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func MapFlagToEnv(cmd *cobra.Command, flagName, envName, configKey string) error {
	// Replace dashes with underscores and convert to uppercase
	err := viper.BindPFlag(configKey, cmd.Flags().Lookup(flagName))
	if err != nil {
		return err
	}
	err = viper.BindEnv(configKey, envName)
	if err != nil {
		return err
	}
	return nil
}

func ConfigureViperConfigFile(cfgFile string) {
	// if --directory is set, use it to set the config file path
	dir := viper.GetString("directory")
	if dir != "" {
		viper.AddConfigPath(dir)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for a config file with the name "config" (without extension).
		viper.AddConfigPath(".ab-config.yaml")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}
}