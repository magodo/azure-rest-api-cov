package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

const azureRepoURL = "git@github.com:Azure/azure-rest-api-specs"
const azureRepoName = "azure-rest-api-specs"

type AzureRepo struct {
	SpecRootDir string
	Repo        *git.Repository
}

// NewAzureRepo will clone the repo from upstream or open the already cloned one from local.
func NewAzureRepo(baseDir string) (*AzureRepo, error) {
	repoDir, err := filepath.Abs(filepath.Join(baseDir, azureRepoName))
	if err != nil {
		return nil, err
	}
	repoDir = filepath.Clean(repoDir)

	if err := os.Mkdir(repoDir, 0755); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("creating repo directory %q: %w", repoDir, err)
		}
	}

	r, err := git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:      azureRepoURL,
		Progress: os.Stdout,
	})
	if err != nil {
		if !errors.Is(err, git.ErrRepositoryAlreadyExists) {
			return nil, fmt.Errorf("cloning azure rest api repo %q: %w", azureRepoURL, err)
		}

		r, err = git.PlainOpen(repoDir)
		if err != nil {
			return nil, fmt.Errorf("open repo %q: %w", repoDir, err)
		}
	}

	repo := AzureRepo{
		SpecRootDir: filepath.Join(repoDir, "specification"),
		Repo:        r,
	}
	return &repo, nil
}

func (repo *AzureRepo) CheckoutRefObj(ref string) (changed bool, err error) {
	return checkoutRef(repo.Repo, ref)
}
