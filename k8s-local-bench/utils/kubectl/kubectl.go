package kubectl

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// Client configures how kubectl is invoked. Optional fields may be nil.
type Client struct {
	// Path to kubectl binary. If nil, will resolve via PATH.
	KubectlPath *string
	// Kubeconfig file path. If nil, uses default kubeconfig behavior.
	Kubeconfig *string
	// Extra args to pass to kubectl (e.g., --namespace)
	ExtraArgs []string
}

// NewClient creates a basic Client.
func NewClient(Kubeconfig *string, ExtraArgs []string) *Client {
	return &Client{
		Kubeconfig: Kubeconfig,
		ExtraArgs:  ExtraArgs,
	}
}

// resolveKubectl returns the binary path to use.
func (c *Client) resolveKubectl() (string, error) {
	if c != nil && c.KubectlPath != nil && *c.KubectlPath != "" {
		return *c.KubectlPath, nil
	}
	p, err := exec.LookPath("kubectl")
	if err != nil {
		return "", fmt.Errorf("kubectl not found in PATH: %w", err)
	}
	return p, nil
}

// buildBaseArgs returns common kubectl args including kubeconfig if set.
func (c *Client) buildBaseArgs() []string {
	var args []string
	if c != nil && c.Kubeconfig != nil && *c.Kubeconfig != "" {
		args = append(args, "--kubeconfig", *c.Kubeconfig)
	}
	if c != nil && len(c.ExtraArgs) > 0 {
		args = append(args, c.ExtraArgs...)
	}
	return args
}

// ApplyPaths takes a list of glob patterns, expands them on the local filesystem,
// and runs `kubectl apply -f` with the matched files and directories. Patterns
// that don't match anything cause an error.
func (c *Client) ApplyPaths(ctx context.Context, patterns []string) error {
	if len(patterns) == 0 {
		return fmt.Errorf("no patterns provided")
	}

	var matches []string
	var unmatched []string
	for _, p := range patterns {
		// If the pattern looks like a URL, pass it through directly
		if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
			matches = append(matches, p)
			continue
		}

		// Expand ~ to home
		if strings.HasPrefix(p, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				p = filepath.Join(home, strings.TrimPrefix(p, "~"))
			}
		}

		g, err := filepath.Glob(p)
		if err != nil {
			return fmt.Errorf("invalid glob pattern %q: %w", p, err)
		}
		if len(g) == 0 {
			unmatched = append(unmatched, p)
			continue
		}
		for _, m := range g {
			matches = append(matches, m)
		}
	}

	if len(unmatched) > 0 {
		return fmt.Errorf("patterns did not match any files: %v", unmatched)
	}

	if len(matches) == 0 {
		return fmt.Errorf("no files to apply")
	}

	kubectlPath, err := c.resolveKubectl()
	if err != nil {
		return err
	}

	// Build command: kubectl apply -f <item1> -f <item2> ... [--kubeconfig ...] [extra args]
	args := []string{"apply"}
	for _, m := range matches {
		args = append(args, "-f", m)
	}
	args = append(args, c.buildBaseArgs()...)

	cmd := exec.CommandContext(ctx, kubectlPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kubectl apply failed: %w", err)
	}
	return nil
}

// ServicePort describes a service port.
type ServicePort struct {
	Name     string
	Port     int
	Protocol string
}

// Service is a simplified representation of a Kubernetes Service useful
// for consumers that only need common fields.
type Service struct {
	Name        string
	Namespace   string
	Type        string
	ClusterIP   string
	ExternalIPs []string
	Ports       []ServicePort
}

// ListServices returns services in the given namespace. If svcType is
// non-nil and non-empty, results are filtered to services whose
// spec.type matches (case-insensitive) the provided value.
func (c *Client) ListServices(ctx context.Context, namespace string, svcType *string) ([]Service, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace must be provided")
	}

	kubectlPath, err := c.resolveKubectl()
	if err != nil {
		return nil, err
	}

	args := []string{"get", "svc", "-n", namespace, "-o", "json"}
	args = append(args, c.buildBaseArgs()...)

	cmd := exec.CommandContext(ctx, kubectlPath, args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("kubectl get services failed: %w", err)
	}

	log.Info().Str("cmd", strings.Join(cmd.Args, " ")).Msg("kubectl command executed")
	log.Info().Str("output", string(out)).Msg("kubectl command output")

	var raw struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
			Spec struct {
				Type        string   `json:"type"`
				ClusterIP   string   `json:"clusterIP"`
				ExternalIPs []string `json:"externalIPs"`
				Ports       []struct {
					Name       string      `json:"name"`
					Port       int         `json:"port"`
					TargetPort interface{} `json:"targetPort"`
					Protocol   string      `json:"protocol"`
				} `json:"ports"`
			} `json:"spec"`
			Status struct {
				LoadBalancer struct {
					Ingress []struct {
						IP string `json:"ip"`
					} `json:"ingress"`
				} `json:"loadBalancer"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse kubectl output: %w", err)
	}

	var svcs []Service
	for _, it := range raw.Items {
		if svcType != nil && *svcType != "" && !strings.EqualFold(it.Spec.Type, *svcType) {
			continue
		}
		s := Service{
			Name:      it.Metadata.Name,
			Namespace: it.Metadata.Namespace,
			Type:      it.Spec.Type,
			ClusterIP: it.Spec.ClusterIP,
		}
		// start with spec.externalIPs (may be nil)
		if len(it.Spec.ExternalIPs) > 0 {
			s.ExternalIPs = append(s.ExternalIPs, it.Spec.ExternalIPs...)
		}
		// include any loadBalancer ingress IPs (status.loadBalancer.ingress[].ip)
		for _, ing := range it.Status.LoadBalancer.Ingress {
			if ing.IP != "" {
				s.ExternalIPs = append(s.ExternalIPs, ing.IP)
			}
		}
		for _, p := range it.Spec.Ports {
			s.Ports = append(s.Ports, ServicePort{Name: p.Name, Port: p.Port, Protocol: p.Protocol})
		}
		svcs = append(svcs, s)
	}

	return svcs, nil
}
