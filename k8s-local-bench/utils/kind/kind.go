package kind

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"
)

// isInstalled returns true if the given executable is found in PATH.
func isInstalled(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// runCmd runs a command and returns combined stdout/stderr.
func runCmd(name string, args ...string) (string, error) {
	log.Debug().Str("cmd", name).Str("args", strings.Join(args, " ")).Msg("running command")
	out, err := exec.Command(name, args...).CombinedOutput()
	return string(out), err
}

// ensureCloudProviderKindInstalled ensures the `cloud-provider-kind` binary exists.
// If missing and `go` is available, it will attempt `go install sigs.k8s.io/cloud-provider-kind@latest`.
func ensureCloudProviderKindInstalled() error {
	if isInstalled("cloud-provider-kind") {
		return nil
	}
	if !isInstalled("go") {
		return fmt.Errorf("go not installed; cannot install cloud-provider-kind")
	}

	out, err := runCmd("go", "install", "sigs.k8s.io/cloud-provider-kind@latest")
	if err != nil {
		return fmt.Errorf("failed to install cloud-provider-kind: %w; output: %s", err, out)
	}
	if !isInstalled("cloud-provider-kind") {
		return fmt.Errorf("cloud-provider-kind not found in PATH after install; output: %s", out)
	}
	log.Info().Msg("cloud-provider-kind installed")
	return nil
}

// Create creates a kind cluster with the provided name. If configPath is non-empty
// it will be passed to `kind create cluster --config`.
func Create(name string, configPath string, kubeconfigPath string) error {
	if !isInstalled("kind") {
		return fmt.Errorf("kind not installed")
	}
	if !isInstalled("docker") {
		return fmt.Errorf("docker not installed")
	}

	// ensure cloud-provider-kind is available (will attempt to install with `go install`)
	if err := ensureCloudProviderKindInstalled(); err != nil {
		return err
	}

	args := []string{"create", "cluster", "--name", name, "--kubeconfig", kubeconfigPath}
	if configPath != "" {
		args = append(args, "--config", configPath)
	}

	out, err := runCmd("kind", args...)
	if err != nil {
		return fmt.Errorf("failed to create kind cluster: %w; output: %s", err, out)
	}
	log.Info().Str("name", name).Msg("kind cluster created")
	return nil
}

// Delete deletes a kind cluster by name.
func Delete(name string) error {
	if !isInstalled("kind") {
		return fmt.Errorf("kind not installed")
	}
	if !isInstalled("docker") {
		return fmt.Errorf("docker not installed")
	}

	out, err := runCmd("kind", "delete", "cluster", "--name", name)
	if err != nil {
		return fmt.Errorf("failed to delete kind cluster: %w; output: %s", err, out)
	}
	log.Info().Str("name", name).Msg("kind cluster deleted")
	return nil
}

// StartLoadBalancer starts the cloud-provider-kind process for the given cluster.
// If background==true the process is started detached and logs are written to
// a temp file; the function returns immediately while the process continues
// running after the CLI exits.
func StartLoadBalancer(clusterName string, background bool) error {
	if err := ensureCloudProviderKindInstalled(); err != nil {
		return err
	}

	args := []string{}

	// determine whether we need sudo
	needSudo := os.Geteuid() != 0
	if needSudo && !isInstalled("sudo") {
		return fmt.Errorf("sudo required but not installed")
	}

	if !background {
		if needSudo {
			// run interactively so user can enter their sudo password
			cmd := exec.Command("sudo", append([]string{"cloud-provider-kind"}, args...)...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("cloud-provider-kind failed: %w", err)
			}
			return nil
		}
		out, err := runCmd("cloud-provider-kind", args...)
		if err != nil {
			return fmt.Errorf("cloud-provider-kind failed: %w; output: %s", err, out)
		}
		return nil
	}

	// background: if sudo is required, first validate sudo credentials interactively
	if needSudo {
		vcmd := exec.Command("sudo", "-v")
		vcmd.Stdout = os.Stdout
		vcmd.Stderr = os.Stderr
		vcmd.Stdin = os.Stdin
		if err := vcmd.Run(); err != nil {
			return fmt.Errorf("sudo validation failed: %w", err)
		}
	}

	// background: start detached with logs redirected to a temp file
	logPath := filepath.Join(os.TempDir(), fmt.Sprintf("cloud-provider-kind-%s.log", clusterName))
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	var cmd *exec.Cmd
	if needSudo {
		cmd = exec.Command("sudo", append([]string{"cloud-provider-kind"}, args...)...)
	} else {
		cmd = exec.Command("cloud-provider-kind", args...)
	}
	cmd.Stdout = f
	cmd.Stderr = f
	cmd.Stdin = nil
	// detach from parent process (Unix)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		f.Close()
		return fmt.Errorf("failed to start cloud-provider-kind: %w", err)
	}
	// we intentionally do not wait; process should keep running after exit
	log.Info().Str("log", logPath).Int("pid", cmd.Process.Pid).Msg("cloud-provider-kind started in background")
	// close our file handle; child keeps file descriptor
	_ = f.Close()
	return nil
}
