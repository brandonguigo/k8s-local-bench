package create

import (
	"context"
	"fmt"
	"localplane/utils/kubectl"
	"time"
)

// WaitForLoadBalancerService polls for a Service of type LoadBalancer in the
// provided namespace and returns the single matched Service. If more than one
// service is present, it keeps waiting until timeout. Returns an error on
// timeout or other failures.
func waitForLoadBalancerService(ctx context.Context, kubeConfig string, namespace string, timeout time.Duration, pollInterval time.Duration) (*kubectl.Service, error) {
	c := kubectl.NewClient(&kubeConfig, nil)
	if c == nil {
		return nil, fmt.Errorf("kubectl client is nil")
	}
	if namespace == "" {
		return nil, fmt.Errorf("namespace must be provided")
	}

	svcType := "LoadBalancer"
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for LoadBalancer service in namespace %s: %w", namespace, ctx.Err())
		case <-ticker.C:
			svcs, err := c.ListServices(ctx, namespace, &svcType)
			if err != nil {
				// transient error; try again until timeout
				continue
			}
			if len(svcs) == 1 {
				// Only return when the single LoadBalancer service has at least one ExternalIP
				if len(svcs[0].ExternalIPs) > 0 {
					return &svcs[0], nil
				}
				// otherwise keep waiting until an ExternalIP is assigned
			}
			// otherwise keep waiting
		}
	}
}
