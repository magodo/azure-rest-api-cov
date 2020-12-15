package workspace

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const azureRepoIndexFilename = "index.json"

type InitOption struct {
	// Force inits workspace from scratch
	Force bool

	// RepoHash specifies the commit hash of azure rest api spec repo to checkout.
	// If this is empty string, then will not modify the repo head after clone.
	RepoHash string
}

func Init(dir string, opt InitOption) (*AzureRepo, *AzureRepoIndex, error) {
	if opt.Force {
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

	var repoChanged bool
	if opt.RepoHash != "" {
		log.Printf("Checkout %s for Azure Repo", opt.RepoHash)
		var err error
		repoChanged, err = repo.CheckoutHash(opt.RepoHash)
		if err != nil {
			return nil, nil, err
		}
	}
	log.Println("Azure Repo ready")

	indexPath := filepath.Join(dir, azureRepoIndexFilename)
	// If repo HEAD changes, we need to regenerate repo index.
	if repoChanged {
		log.Println("Remove Azure Repo Index (if any) since repo HEAD changes")
		os.Remove(indexPath)
	}

	index, err := ImportAzureRepoIndex(indexPath)
	if err == nil {
		return repo, index, nil
	}

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
	return repo, index, nil
}
