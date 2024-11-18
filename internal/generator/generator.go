package generator

import (
	"fmt"
	"path/filepath"

	"github.com/belingud/gptcommit/internal/config"
	"github.com/belingud/gptcommit/internal/git"
	"github.com/belingud/gptcommit/internal/llm"
)

// Generator handles commit message generation
type Generator struct {
	config *config.Config
	repo   *git.Repository
	llm    *llm.Client
}

// NewGenerator creates a new message generator
func NewGenerator(cfg *config.Config, repoPath string) (*Generator, error) {
	repo, err := git.NewRepository(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create git repository: %w", err)
	}

	client, err := llm.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	return &Generator{
		config: cfg,
		repo:   repo,
		llm:    client,
	}, nil
}

// shouldIgnoreFile checks if a file should be ignored
func (g *Generator) shouldIgnoreFile(file string) (bool, error) {
	// Check if file is in git ignore
	ignored, err := g.repo.IsIgnored(file)
	if err != nil {
		return false, err
	}
	if ignored {
		return true, nil
	}

	// Check if file matches any patterns in config.FileIgnore
	for _, pattern := range g.config.FileIgnore {
		matched, err := filepath.Match(pattern, file)
		if err != nil {
			return false, fmt.Errorf("invalid pattern %s: %w", pattern, err)
		}
		if matched {
			return true, nil
		}
	}

	return false, nil
}

// filterIgnoredFiles removes ignored files from the list
func (g *Generator) filterIgnoredFiles(files []string) ([]string, error) {
	var result []string
	for _, file := range files {
		ignore, err := g.shouldIgnoreFile(file)
		if err != nil {
			return nil, err
		}
		if !ignore {
			result = append(result, file)
		}
	}
	return result, nil
}

// GenerateMessage generates a commit message for staged changes
func (g *Generator) GenerateMessage() (string, error) {
	// Check if there are staged changes
	hasChanges, err := g.repo.HasStagedChanges()
	if err != nil {
		return "", err
	}
	if !hasChanges {
		return "", fmt.Errorf("no staged changes")
	}

	// Get staged files
	files, err := g.repo.GetStagedFiles()
	if err != nil {
		return "", err
	}

	// Filter out ignored files
	files, err = g.filterIgnoredFiles(files)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no non-ignored files staged")
	}

	// Get diff
	diff, err := g.repo.GetStagedDiff()
	if err != nil {
		return "", err
	}

	// Generate commit message
	message, err := g.llm.GenerateCommitMessage(diff)
	if err != nil {
		return "", err
	}

	return message, nil
}

// CommitChanges generates a message and commits the changes
func (g *Generator) CommitChanges() error {
	message, err := g.GenerateMessage()
	if err != nil {
		return err
	}

	return g.repo.Commit(message)
}
