package workspace

import (
	"encoding/hex"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

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
		Progress: os.Stdout,
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
