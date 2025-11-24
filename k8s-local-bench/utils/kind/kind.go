package kind

import (
	"fmt"
	"os/exec"
	"strings"

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

// Create creates a kind cluster with the provided name. If configPath is non-empty
// it will be passed to `kind create cluster --config`.
func Create(name string, configPath string) error {
	if !isInstalled("kind") {
		return fmt.Errorf("kind not installed")
	}
	if !isInstalled("docker") {
		return fmt.Errorf("docker not installed")
	}

	args := []string{"create", "cluster", "--name", name}
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
