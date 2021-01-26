package workspace

import (
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/go-git/go-git/v5"
)

type GrantRepo struct {
	RootDir string
	Repo    *git.Repository
}

func NewGrantRepo(baseDir, repoUrl string, auth transport.AuthMethod) (*GrantRepo, error) {
	repoName := "grant"
	repoDir, err := filepath.Abs(filepath.Join(baseDir, repoName))
	if err != nil {
		return nil, err
	}
	repoDir = filepath.Clean(repoDir)

	r, err := cloneAndOpenRepo(repoDir, repoUrl, auth)
	if err != nil {
		return nil, err
	}

	return &GrantRepo{
		RootDir: repoDir,
		Repo:    r,
	}, nil
}

func (repo *GrantRepo) CheckoutRef(ref string, auth transport.AuthMethod) (changed bool, err error) {
	return checkoutRef(repo.Repo, ref, auth)
}
