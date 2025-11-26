/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package clusterCmd

import (
	"localplane/cmd/cluster/create"
	"localplane/cmd/cluster/destroy"

	"github.com/spf13/cobra"
)

// NewCommand creates the cluster command
func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "cluster",
		Short: "control local k8s clusters",
	}

	cmd.PersistentFlags().String("cluster-name", "", "name of the cluster (directory under CLI config clusters/)")

	// add subcommands here
	cmd.AddCommand(create.NewCommand())
	cmd.AddCommand(destroy.NewCommand())
	return cmd
}
