package workspace

import (
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

const azureRepoURL = "https://github.com/Azure/azure-rest-api-specs.git"
const azureRepoName = "azure-rest-api-specs"

type AzureRepo struct {
	SpecRootDir string
	Repo        *git.Repository
}

func NewAzureRepo(baseDir string) (*AzureRepo, error) {
	repoDir, err := filepath.Abs(filepath.Join(baseDir, azureRepoName))
	if err != nil {
		return nil, err
	}
	repoDir = filepath.Clean(repoDir)

	r, err := cloneAndOpenRepo(repoDir, azureRepoURL, nil)
	if err != nil {
		return nil, err
	}

	return &AzureRepo{
		SpecRootDir: filepath.Join(repoDir, "specification"),
		Repo:        r,
	}, nil
}

func (repo *AzureRepo) CheckoutRef(ref string) (changed bool, err error) {
	return checkoutRef(repo.Repo, ref, nil)
}
