package workspace

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// cloneAndOpenRepo will clone the repo from upstream if not existed, then open it.
func cloneAndOpenRepo(repoDir string, repoUrl string, auth transport.AuthMethod) (*git.Repository, error) {
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
			Auth:     auth,
			Progress: os.Stderr,
		}); err != nil {
			os.RemoveAll(repoDir)
			return nil, fmt.Errorf("cloning repo %q: %w", repoUrl, err)

		}
	}

	r, err := git.PlainOpen(repoDir)
	if err != nil {
		return nil, fmt.Errorf("open repo %q: %w", repoDir, err)
	}

	return r, nil
}

func checkoutRef(repo *git.Repository, ref string, auth transport.AuthMethod) (changed bool, err error) {
	wt, err := repo.Worktree()
	if err != nil {
		return false, err
	}
	ohead, err := repo.Head()
	if err != nil {
		return false, err
	}

	if err := repo.Fetch(&git.FetchOptions{
		Auth:     auth,
		Progress: os.Stderr,
		Force:    true,
	}); err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return false, err
		}
	}

	opt := &git.CheckoutOptions{Force: true}
	if _, err := hex.DecodeString(ref); err == nil {
		opt.Hash = plumbing.NewHash(ref)
		if err := wt.Checkout(opt); err != nil {
			return false, err
		}
	} else {
		opt.Branch = plumbing.NewTagReferenceName(ref)
		if err := wt.Checkout(opt); err != nil {
			if !errors.Is(err, plumbing.ErrReferenceNotFound) {
				return false, err
			}

			opt.Branch = plumbing.NewRemoteReferenceName("origin", ref)
			if err := wt.Checkout(opt); err != nil {
				return false, err
			}
		}
	}
	head, _ := repo.Head()
	return head.Hash() == ohead.Hash(), nil
}
