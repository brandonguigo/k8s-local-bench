package dnsmasq

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Client manages dnsmasq configuration updates.
type Client struct {
	// ConfigPath is the dnsmasq config file to edit. If empty, a sensible
	// default will be chosen from common locations.
	ConfigPath string
}

// NewClient creates a dnsmasq Client. Pass an empty string to use defaults.
func NewClient(configPath string) *Client {
	return &Client{ConfigPath: configPath}
}

// EnsureDomainIP ensures dnsmasq has an `address=/domain/ip` entry for the
// provided domain. If an entry exists it will be replaced; otherwise it will
// be appended. After writing the config the function attempts to reload
// dnsmasq (best-effort).
func (c *Client) EnsureDomainIP(ctx context.Context, domain, ip string) error {
	if domain == "" {
		return fmt.Errorf("domain must be provided")
	}
	if ip == "" {
		return fmt.Errorf("ip must be provided")
	}

	// verify dnsmasq is available
	if _, err := exec.LookPath("dnsmasq"); err != nil {
		return fmt.Errorf("dnsmasq not found in PATH: %w", err)
	}

	// determine config path
	cfg := c.ConfigPath
	candidates := []string{
		"/opt/homebrew/etc/dnsmasq.conf",
		"/usr/local/etc/dnsmasq.conf",
		"/etc/dnsmasq.conf",
	}
	if cfg == "" {
		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				cfg = p
				break
			}
		}
		if cfg == "" {
			// default to first candidate if none exist
			cfg = candidates[0]
		}
	}

	// read existing file if present
	var content string
	if b, err := os.ReadFile(cfg); err == nil {
		content = string(b)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read dnsmasq config %s: %w", cfg, err)
	}

	lines := []string{}
	if content != "" {
		lines = strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	}

	wantPrefix := fmt.Sprintf("address=/%s/", domain)
	wantLine := wantPrefix + ip

	replaced := false
	for i, l := range lines {
		// ignore leading/trailing whitespace when matching
		tl := strings.TrimSpace(l)
		if strings.HasPrefix(tl, "#") {
			continue
		}
		if strings.HasPrefix(tl, wantPrefix) {
			lines[i] = wantLine
			replaced = true
			break
		}
	}
	if !replaced {
		lines = append(lines, wantLine)
	}

	// ensure directory exists
	if err := os.MkdirAll(filepath.Dir(cfg), 0o755); err != nil {
		return fmt.Errorf("failed to create dnsmasq config dir: %w", err)
	}

	// write atomically
	tmp, err := os.CreateTemp(filepath.Dir(cfg), "dnsmasq.conf.tmp.*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := tmp.WriteString(strings.Join(lines, "\n") + "\n"); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to write temp config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp config: %w", err)
	}

	if err := os.Rename(tmpPath, cfg); err != nil {
		return fmt.Errorf("failed to move temp config into place: %w", err)
	}

	// attempt to reload dnsmasq: prefer `brew services restart dnsmasq` if brew exists
	if _, err := exec.LookPath("brew"); err == nil {
		// best-effort; ignore output but return error if command fails
		cmd := exec.CommandContext(ctx, "brew", "services", "restart", "dnsmasq")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to restart dnsmasq via brew: %w; output: %s", err, string(out))
		}
		return nil
	}

	// fallback: send HUP to dnsmasq processes
	if _, err := exec.LookPath("pkill"); err == nil {
		cmd := exec.CommandContext(ctx, "pkill", "-HUP", "dnsmasq")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to HUP dnsmasq with pkill: %w; output: %s", err, string(out))
		}
		return nil
	}

	// last resort: try killall
	if _, err := exec.LookPath("killall"); err == nil {
		cmd := exec.CommandContext(ctx, "killall", "-HUP", "dnsmasq")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to HUP dnsmasq with killall: %w; output: %s", err, string(out))
		}
		return nil
	}

	// if we reach here we wrote the config but couldn't reload automatically
	return fmt.Errorf("updated dnsmasq config at %s but could not reload dnsmasq (no reload command available)", cfg)
}
