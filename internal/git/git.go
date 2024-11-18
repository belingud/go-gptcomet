package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repository represents a Git repository
type Repository struct {
	path string
}

// NewRepository creates a new Git repository instance
func NewRepository(path string) (*Repository, error) {
	// If path is empty, use current directory
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Check if path is a git repository
	gitPath := filepath.Join(path, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s is not a git repository", path)
	}

	return &Repository{path: path}, nil
}

// GetStagedDiff returns the diff of staged changes
func (r *Repository) GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	cmd.Dir = r.path

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// GetStagedFiles returns a list of staged files
func (r *Repository) GetStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	cmd.Dir = r.path

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w, stderr: %s", err, stderr.String())
	}

	files := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(files) == 1 && files[0] == "" {
		return nil, nil
	}

	return files, nil
}

// Commit creates a new commit with the given message
func (r *Repository) Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = r.path

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

// HasStagedChanges checks if there are any staged changes
func (r *Repository) HasStagedChanges() (bool, error) {
	files, err := r.GetStagedFiles()
	if err != nil {
		return false, err
	}

	return len(files) > 0, nil
}

// IsIgnored checks if a file is ignored by git
func (r *Repository) IsIgnored(file string) (bool, error) {
	cmd := exec.Command("git", "check-ignore", file)
	cmd.Dir = r.path

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// Exit code 1 means the file is not ignored
			return false, nil
		}
		return false, fmt.Errorf("failed to check if file is ignored: %w", err)
	}

	// Exit code 0 means the file is ignored
	return true, nil
}
