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
	WorkspaceDir string

	// Force inits workspace from scratch
	Force bool

	// RegenIndex regenerates the index
	RegenIndex bool

	// SpecRepoRef specifies the git ref object of azure rest api spec repo to check out.
	// If this is empty string, then will not modify the repo head after clones.
	SpecRepoRef string

	GrantRepoURL string
	// GrantRepoRef specifies the git ref object of grant repo to check out.
	// If this is empty string, then will not modify the repo head after clones.
	GrantRepoRef string
}

type Workspace struct {
	AzureRepo      *AzureRepo
	AzureRepoIndex *AzureRepoIndex
	GrantRepo      *GrantRepo
}

func Init(opt InitOption) (*Workspace, error) {
	dir := opt.WorkspaceDir
	if opt.Force {
		if err := os.RemoveAll(dir); err != nil {
			return nil, fmt.Errorf("cleaning up the workspace %q: %w", dir, err)
		}
	}

	var result Workspace
	var err error
	log.Printf("Initializing workspace under: %s", dir)
	if err := os.Mkdir(dir, 0755); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("creating workspace %q: %w", dir, err)
		}
	}

	// Initialize grant repository.
	if opt.GrantRepoURL != "" {
		log.Printf("Preparing Granting Repo %q...", opt.GrantRepoURL)
		result.GrantRepo, err = NewGrantRepo(dir, opt.GrantRepoURL)
		if err != nil {
			return nil, err
		}

		if opt.GrantRepoRef != "" {
			log.Printf("Checkout %s for granting repo", opt.GrantRepoRef)
			if _, err := result.GrantRepo.CheckoutRef(opt.GrantRepoRef); err != nil {
				return nil, err
			}
		}
	}

	// Initialize azure repository.
	log.Println("Preparing Azure Repo...")
	repo, err := NewAzureRepo(dir)
	if err != nil {
		return nil, err
	}
	result.AzureRepo = repo

	var repoChanged bool
	if opt.SpecRepoRef != "" {
		log.Printf("Checkout %s for Azure Repo", opt.SpecRepoRef)
		var err error
		repoChanged, err = repo.CheckoutRefObj(opt.SpecRepoRef)
		if err != nil {
			return nil, err
		}
	}
	headRef, _ := repo.Repo.Head()
	log.Println("Azure Repo ready")

	indexPath := filepath.Join(dir, azureRepoIndexFilename)

	// If repo HEAD changes, we need to regenerate repo index.
	if repoChanged || opt.RegenIndex {
		log.Println("Remove Azure Repo Index (if any) since repo HEAD changes")
		os.Remove(indexPath)
	}

	index, err := ImportAzureRepoIndex(indexPath)
	if err == nil {
		result.AzureRepoIndex = index
		return &result, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	// Index not exported yet, need to generate the index and export it then.
	log.Println("Generating Azure Repo Index...")
	index, err = NewAzureRepoIndex(repo.SpecRootDir, headRef.Hash().String())
	if err != nil {
		return nil, err
	}
	result.AzureRepoIndex = index
	log.Println("Azure Repo Index generated")

	if err := index.Export(indexPath); err != nil {
		return nil, err
	}
	log.Println("Azure Repo Index exported")

	return &result, nil
}
