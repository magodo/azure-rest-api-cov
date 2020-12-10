package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
)

const azureRepoURL = "https://github.com/Azure/azure-rest-api-specs.git"

type AzureRepo struct {
	SpecRootDir string
	Repo        *git.Repository
}

func NewAzureRepo(repoDir string) (*AzureRepo, error) {
	repoDir, err := filepath.Abs(repoDir)
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
	if err := repo.Pull(); err != nil {
		return nil, err
	}
	return &repo, nil
}

func (repo *AzureRepo) Pull() error {
	wt, err := repo.Repo.Worktree()
	if err != nil {
		return err
	}

	if err := wt.Pull(&git.PullOptions{
		Progress: os.Stdout,
	}); err != nil {
		if err != git.NoErrAlreadyUpToDate {
			return err
		}
	}
	return nil
}

func (repo AzureRepo) GenerateIndex() (*AzureRepoIndex, error) {
	index := &AzureRepoIndex{
		SpecRootDir:       repo.SpecRootDir,
		ResourceProviders: map[ResourceProvider]APIs{},
	}
	schemaFolderPattern := regexp.MustCompile(fmt.Sprintf(`^(\w+(%[1]sresource-manager(%[1]sMicrosoft.\w+(%[1]s(preview|stable)(%[1]s([^%[1]s]+))?)?)?)?)?$`, regexp.QuoteMeta(string(os.PathSeparator))))
	schemaPattern := regexp.MustCompile(fmt.Sprintf(`^(\w+)%[1]sresource-manager%[1]sMicrosoft.\w+%[1]s(preview|stable)%[1]s(.+?)%[1]s.+?\.json$`, regexp.QuoteMeta(string(os.PathSeparator))))
	if err := filepath.Walk(repo.SpecRootDir,
		func(p string, info os.FileInfo, err error) error {
			relPath := strings.TrimPrefix(strings.TrimPrefix(p, repo.SpecRootDir), string(os.PathSeparator))
			if err != nil {
				return err
			}
			if info == nil {
				return nil
			}

			// Skip directories not match the schema folder patterns
			if info.IsDir() {
				if !schemaFolderPattern.MatchString(relPath) {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip files not match the schema file patterns
			match := schemaPattern.FindStringSubmatch(relPath)
			if len(match) == 0 {
				return nil
			}

			rpName, apiState, apiVersion := match[1], match[2], match[3]
			api := API{
				Version:   apiVersion,
				IsPreview: apiState == "preview",
			}

			if err := index.recordSpec(relPath, rpName, api); err != nil {
				return err
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return index, nil
}
