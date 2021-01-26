package workspace

import (
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

type GrantRepo struct {
	RootDir string
	Repo    *git.Repository
}

func NewGrantRepo(baseDir string, repoUrl string) (*GrantRepo, error) {
	repoName := "grant"
	repoDir, err := filepath.Abs(filepath.Join(baseDir, repoName))
	if err != nil {
		return nil, err
	}
	repoDir = filepath.Clean(repoDir)

	r, err := cloneAndOpenRepo(repoDir, repoUrl)
	if err != nil {
		return nil, err
	}

	return &GrantRepo{
		RootDir: repoDir,
		Repo:    r,
	}, nil
}

func (repo *GrantRepo) CheckoutRef(ref string) (changed bool, err error) {
	return checkoutRef(repo.Repo, ref)
}
