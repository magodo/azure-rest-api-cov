package swagger

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

type SpecLocator struct {
	ApiVersion string `json:"api_version"`
	ApiPath    string `json:"api_path"`
}

type SpecLookupTables map[string]SpecLookupTable

var specLookupTables = map[string]SpecLookupTable{}

func LookupSpecPath(swaggerRepoSpecBasePath string, locator SpecLocator) (string, error) {
	t, ok := specLookupTables[swaggerRepoSpecBasePath]
	if !ok {
		t = SpecLookupTable{
			m:                       map[SpecLocator]string{},
			isInit:                  false,
			swaggerRepoSpecBasePath: swaggerRepoSpecBasePath,
		}
	}
	return t.Lookup(locator)
}

type SpecLookupTable struct {
	m                       map[SpecLocator]string // maps from spec locator to the spec rel-path.
	isInit                  bool
	swaggerRepoSpecBasePath string
}

func (t *SpecLookupTable) Lookup(locator SpecLocator) (string, error) {
	if !t.isInit {
		if err := t.initOnce(); err != nil {
			return "", err
		}
	}
	path, ok := t.m[locator]
	if !ok {
		return "", fmt.Errorf("no spec of version %q contains path %q", locator.ApiVersion, locator.ApiPath)
	}
	return path, nil
}

func (t *SpecLookupTable) initOnce() error {
	var m sync.Map
	g := new(errgroup.Group)
	g.Go(func() error {
		schemaFolderPattern := regexp.MustCompile(fmt.Sprintf(`^(%[1]s\w+(%[1]sresource-manager(%[1]sMicrosoft.\w+(%[1]s(preview|stable)(%[1]s([^%[1]s]+))?)?)?)?)?$`, regexp.QuoteMeta(string(os.PathSeparator))))
		schemaPattern := regexp.MustCompile(fmt.Sprintf(`^%[1]s\w+%[1]sresource-manager%[1]sMicrosoft.\w+%[1]s(preview|stable)%[1]s.+?%[1]s.+?\.json$`, regexp.QuoteMeta(string(os.PathSeparator))))
		if err := filepath.Walk(t.swaggerRepoSpecBasePath,
			func(p string, info os.FileInfo, err error) error {
				p, _ = filepath.Abs(p)
				t.swaggerRepoSpecBasePath, _ = filepath.Abs(t.swaggerRepoSpecBasePath)
				relPath := strings.TrimPrefix(p, t.swaggerRepoSpecBasePath)
				if err != nil {
					return err
				}
				if info == nil {
					return nil
				}

				// Skip directories not match the schema folder patterns
				if info.IsDir() {
					if !schemaFolderPattern.MatchString(relPath) {
						//log.Printf("Skip directory %s!\n", relPath)
						return filepath.SkipDir
					}
					return nil
				}

				// Skip files not match the schema file patterns
				if !schemaPattern.MatchString(relPath) {
					//log.Printf("Skip file %s!\n", relPath)
					return nil
				}

				spec, err := LoadSwagger(t.swaggerRepoSpecBasePath + relPath)
				//log.Printf("Loading %s\n", relPath)
				if err != nil {
					return err
				}
				if spec.Paths == nil {
					return nil
				}

				specPath := strings.TrimLeft(relPath, string(os.PathSeparator))
				for p := range spec.Paths.Paths {
					m.Store(SpecLocator{
						ApiPath:    p,
						ApiVersion: spec.Info.Version,
					}, specPath)
				}

				return nil
			}); err != nil {
			return err
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return err
	}

	m.Range(func(key, value interface{}) bool {
		t.m[key.(SpecLocator)] = value.(string)
		return true
	})
	return nil
}
