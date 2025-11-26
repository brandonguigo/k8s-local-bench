package argocd

import (
	"errors"
	"fmt"
	"k8s-local-bench/config"
	stdlog "log"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// RepoMount defines a name/hostPath/mountPath triple for mounting a repository
// into Argo CD's repo-server.
type RepoMount struct {
	Name      string
	HostPath  string
	MountPath string
}

// Client is a small helper to configure operations that may need common
// configuration such as a kubeconfig path.
type Client struct {
	Kubeconfig string
}

// NewClient creates a configured Client. Pass empty string for defaults.
func NewClient(kubeconfig string) *Client {
	return &Client{Kubeconfig: kubeconfig}
}

// InstallOrUpgradeArgoCD installs or upgrades ArgoCD using the Helm SDK (upgrade --install).
// - mounts: list of RepoMount to add to repoServer.volumes and repoServer.volumeMounts
func (c *Client) InstallOrUpgradeArgoCD(mounts []RepoMount) (string, error) {
	// use official argo-cd chart from Argo Helm
	release := "argocd"
	namespace := "argocd"
	repoURL := "https://argoproj.github.io/argo-helm"
	chart := "argo-cd"
	// prepare values map
	values := map[string]interface{}{}
	repoServer := map[string]interface{}{}
	server := map[string]interface{}{}

	vols := []interface{}{}
	vms := []interface{}{}

	global := map[string]interface{}{}
	configs := map[string]interface{}{}

	for _, m := range mounts {
		vols = append(vols, map[string]interface{}{
			"name": m.Name,
			"hostPath": map[string]interface{}{
				"path": m.HostPath,
				"type": "Directory",
			},
		})
		vms = append(vms, map[string]interface{}{
			"name":      m.Name,
			"mountPath": m.MountPath,
		})
	}
	repoServer["volumes"] = vols
	repoServer["volumeMounts"] = vms

	// add an ingress with the local dnsmasq domain (argocd.k8s-bench.local)
	host := "argocd.k8s-bench.local"
	global["domain"] = host
	configs["params"] = map[string]interface{}{
		"server.insecure": "true",
	}

	// configure access without login
	configs["cm"] = map[string]interface{}{
		"users.anonymous.enabled": "true",
	}
	configs["rbac"] = map[string]interface{}{
		"policy.default": "role:admin",
	}

	server["ingress"] = map[string]interface{}{
		"enabled":          true,
		"ingressClassName": "haproxy",
		"tls":              false,
	}

	values["global"] = global
	values["configs"] = configs
	values["repoServer"] = repoServer
	values["server"] = server

	// helm SDK config
	settings := cli.New()
	if c != nil && c.Kubeconfig != "" {
		settings.KubeConfig = c.Kubeconfig
	}
	var cfg action.Configuration
	var helmOutput = func(format string, v ...interface{}) { /* no-op */ }
	if config.CliConfig.Debug {
		helmOutput = stdlog.Printf
	}
	if err := cfg.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), helmOutput); err != nil {
		return "", fmt.Errorf("failed to init helm configuration: %w", err)
	}

	// locate and load chart (supports repo URL via ChartPathOptions)
	cp := action.ChartPathOptions{RepoURL: repoURL}
	chartPath, err := cp.LocateChart(chart, settings)
	if err != nil {
		return "", fmt.Errorf("locate chart: %w", err)
	}
	ch, err := loader.Load(chartPath)
	if err != nil {
		return "", fmt.Errorf("load chart: %w", err)
	}

	// Prepare variables to capture release name and version after install/upgrade.
	var relName string
	var relVersion int
	// If not found -> install, else upgrade.
	g := action.NewGet(&cfg)
	_, err = g.Run(release)
	if err != nil {
		// If the release is not found, perform an install.
		if errors.Is(err, driver.ErrReleaseNotFound) {
			i := action.NewInstall(&cfg)
			i.ReleaseName = release
			i.Namespace = namespace
			i.CreateNamespace = true
			i.Timeout = time.Duration(5) * time.Minute
			i.Wait = true
			rel, err := i.Run(ch, values)
			if err != nil {
				return "", fmt.Errorf("install failed: %w", err)
			}
			relName = rel.Name
			relVersion = rel.Version
		} else {
			return "", fmt.Errorf("failed checking release: %w", err)
		}
	} else {
		// Release exists -> perform upgrade.
		u := action.NewUpgrade(&cfg)
		u.Namespace = namespace
		u.Wait = true
		rel, err := u.Run(release, ch, values)
		if err != nil {
			return "", fmt.Errorf("upgrade failed: %w", err)
		}
		relName = rel.Name
		relVersion = rel.Version
	}

	log.Info().Str("release", relName).Int("version", relVersion).Str("namespace", namespace).Msg("argocd installed/updated via helm sdk")
	return fmt.Sprintf("release %s (version %d)", relName, relVersion), nil
}
