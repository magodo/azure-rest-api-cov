package workspace

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// cloneAndOpenRepo will clone the repo from upstream if not existed, then open it.
func cloneAndOpenRepo(repoDir string, repoUrl string) (*git.Repository, error) {
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
		// URLs must have at least one element, which will be used for fetching.
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
			Progress: os.Stderr,
		}); err != nil {
			os.RemoveAll(repoDir)
			return nil, fmt.Errorf("cloning repo %q: %w", azureRepoURL, err)

		}
	}

	r, err := git.PlainOpen(repoDir)
	if err != nil {
		return nil, fmt.Errorf("open repo %q: %w", repoDir, err)
	}

	return r, nil
}

func checkoutRef(repo *git.Repository, ref string) (changed bool, err error) {
	wt, err := repo.Worktree()
	if err != nil {
		return false, err
	}
	head, err := repo.Head()
	if err != nil {
		return false, err
	}

	var isHash bool
	if _, err := hex.DecodeString(ref); err == nil {
		isHash = true
	} else {
		ref = string(plumbing.NewBranchReferenceName(ref))
	}

	// Nothing to do since already in the ref
	if isHash {
		if head.Hash().String() == ref {
			return false, nil
		}
	} else {
		if string(head.Name()) == ref {
			return false, nil
		}
	}

	// Fetch from default remote
	if err := wt.Pull(&git.PullOptions{
		Progress: os.Stderr,
	}); err != nil {
		if err != git.NoErrAlreadyUpToDate {
			return false, err
		}
	}

	opt := &git.CheckoutOptions{Force: true}
	if isHash {
		opt.Hash = plumbing.NewHash(ref)
	} else {
		opt.Branch = plumbing.ReferenceName(ref)
	}

	// Checkout the ref
	if err := wt.Checkout(opt); err != nil {
		return false, err
	}
	return true, nil
}
