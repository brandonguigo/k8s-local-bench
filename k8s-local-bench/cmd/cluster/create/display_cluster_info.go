package create

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

func displayClusterInfo(clusterName, argoCDUrl string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Service", "Info"})
	t.AppendRows([]table.Row{
		{"Cluster Name", clusterName},
		{"ArgoCD URL", argoCDUrl},
	})
	t.Render()
}
