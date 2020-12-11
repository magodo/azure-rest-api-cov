package workspace

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const azureRepoIndexFilename = "index.json"

func Init(dir string, force bool) (*AzureRepo, *AzureRepoIndex, error) {
	if force {
		if err := os.RemoveAll(dir); err != nil {
			return nil, nil, fmt.Errorf("cleaning up the workspace %q: %w", dir, err)
		}
	}

	log.Printf("Initializing workspace under: %s", dir)
	if err := os.Mkdir(dir, 0755); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return nil, nil, fmt.Errorf("creating workspace %q: %w", dir, err)
		}
	}

	// Initialize azure repository.
	log.Println("Preparing Azure Repo...")
	repo, err := NewAzureRepo(dir)
	if err != nil {
		return nil, nil, err
	}
	log.Println("Pull origin/main for Azure Repo...")
	if err := repo.Pull(); err != nil {
		return nil, nil, err
	}
	log.Println("Azure Repo ready")

	// Try to load index file from fs
	indexPath := filepath.Join(dir, azureRepoIndexFilename)
	index, err := ImportAzureRepoIndex(indexPath)
	if err == nil {
		log.Println("Refresh Azure Repo Index")
		// For imported index, we should refresh it to ensure it is up-to-date
		if err := index.Refresh(); err != nil {
			return nil, nil, err
		}
	} else {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, nil, err
		}

		// Index not exported yet, need to generate the index and export it then.
		log.Println("Generating Azure Repo Index...")
		index, err = NewAzureRepoIndex(repo.SpecRootDir)
		if err != nil {
			return nil, nil, err
		}
		log.Println("Azure Repo Index generated")
		if err := index.Export(indexPath); err != nil {
			return nil, nil, err
		}
		log.Println("Azure Repo Index exported")
	}
	return repo, index, nil
}
