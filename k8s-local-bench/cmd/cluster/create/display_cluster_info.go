package create

import (
	"fmt"
)

func displayClusterInfo(clusterName, kubeconfigPath, argoCDUrl, headlampUrl, headlampSecret string) {
	fmt.Println()
	fmt.Println()

	fmt.Printf("🎉 Cluster '%s' created successfully! 🎉", clusterName)
	fmt.Println()
	fmt.Println()

	fmt.Printf("Access your cluster services at the following URLs:")
	fmt.Println()

	fmt.Printf("🗂️ Kubeconfig: %s", kubeconfigPath)
	fmt.Println()
	fmt.Printf("🥷🏻 ArgoCD:   http://%s", argoCDUrl)
	fmt.Println()
	fmt.Printf("🔍 Headlamp: http://%s", headlampUrl)
	fmt.Println()
	fmt.Printf("🔑 Headlamp Token: %s", headlampSecret)
	fmt.Println()

}
