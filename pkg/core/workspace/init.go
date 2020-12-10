package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const repoName = "azure-rest-api-specs"

type InitOption struct {
	force bool
}

func Init(dir string, opt *InitOption) (*AzureRepo, error) {
	if opt != nil && opt.force {
		if err := os.RemoveAll(dir); err != nil {
			return nil, fmt.Errorf("cleaning up the workspace %q: %w", dir, err)
		}
	}

	if err := os.Mkdir(dir, 0755); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("creating workspace %q: %w", dir, err)
		}
	}

	// Initialize azure repository.
	azureRepoDir := filepath.Join(dir, repoName)
	repo, err := NewAzureRepo(azureRepoDir)
	_, _ = repo, err

	// Try to load index file from fs

	return repo, nil
}
