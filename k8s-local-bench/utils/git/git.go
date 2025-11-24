package git

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// InitializeGitRepo ensures the target path is an empty directory (creates it if missing)
// and initializes an empty git repository there. It does not add any remotes/origins.
func InitializeGitRepo(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("path %s exists and is not a directory", path)
		}
		// check directory is empty
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("opening directory %s: %w", path, err)
		}
		defer f.Close()
		// Try to read a single entry; if EOF -> empty
		_, err = f.Readdirnames(1)
		if err != nil && err != io.EOF {
			return fmt.Errorf("reading directory %s: %w", path, err)
		}
		if err != io.EOF {
			return fmt.Errorf("directory %s is not empty", path)
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", path, err)
		}
	} else {
		return fmt.Errorf("stat %s: %w", path, err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git init failed: %v: %s", err, string(out))
	}

	return nil
}
