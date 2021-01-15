package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

type GrantRepo struct {
	RootDir string
}

func NewGrantRepo(baseDir string, repoUrl string) (*GrantRepo, error) {
	repoName := "grant"
	repoDir, err := filepath.Abs(filepath.Join(baseDir, repoName))
	if err != nil {
		return nil, err
	}
	repoDir = filepath.Clean(repoDir)

	var repoExist bool
	if _, err := os.Stat(repoDir); err == nil {
		r, err := git.PlainOpen(repoDir)
		if err != nil {
			return nil, err
		}
		remote, err := r.Remote("origin")
		if err != nil {
			return nil, err
		}
		// URLs must have at least one element, which will be used by fetch.
		if remote.Config().URLs[0] != repoUrl {
			os.RemoveAll(repoDir)
		} else {
			repoExist = true
		}
	}

	if !repoExist {
		if err := os.Mkdir(repoDir, 0755); err != nil {
			if !errors.Is(err, os.ErrExist) {
				return nil, fmt.Errorf("creating repo directory %q: %w", repoDir, err)
			}
		}

		if _, err := git.PlainClone(repoDir, false, &git.CloneOptions{
			URL:      repoUrl,
			Progress: os.Stdout,
		}); err != nil {
			return nil, fmt.Errorf("cloning grant repo %q: %w", azureRepoURL, err)

		}
	}

	r, err := git.PlainOpen(repoDir)
	if err != nil {
		return nil, fmt.Errorf("open repo %q: %w", repoDir, err)
	}

	// Always use the latest version of grant repo
	wt, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	// Fetch from default remote
	if err := wt.Pull(&git.PullOptions{
		Progress: os.Stdout,
	}); err != nil {
		if err != git.NoErrAlreadyUpToDate {
			return nil, err
		}
	}
	head, err := r.Head()
	if err != nil {
		return nil, err
	}
	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: head.Name(),
		Force:  true,
	}); err != nil {
		return nil, err
	}

	repo := GrantRepo{
		RootDir: repoDir,
	}

	return &repo, nil
}
