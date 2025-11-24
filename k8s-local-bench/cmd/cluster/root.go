/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package clusterCmd

import (
	"k8s-local-bench/cmd/cluster/create"

	"github.com/spf13/cobra"
)

// NewCommand creates the cluster command
func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "cluster",
		Short: "control local k8s clusters",
	}

	// flag to allow specifying a kind config file; if empty we'll look in CWD
	cmd.PersistentFlags().StringP("kind-config", "k", "", "path to kind config file (searched in current directory if unspecified)")
	// add subcommands here
	cmd.AddCommand(create.NewCommand())
	return cmd
}
