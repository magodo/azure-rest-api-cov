package swagger

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/magodo/azure-rest-api-cov/pkg/core/swaggerloader"

	"golang.org/x/sync/errgroup"
)

type SpecLocator struct {
	ApiVersion string `json:"api_version"`
	ApiPath    string `json:"api_path"`
}

type SpecLookupTables map[string]SpecLookupTable

var specLookupTables = map[string]SpecLookupTable{}

func (ts SpecLookupTables) Init() error {
	return nil
}

func LookupSpecPath(specBasePath string, locator SpecLocator) (string, error) {
	p, err := filepath.Abs(specBasePath)
	if err != nil {
		return "", err
	}
	p = filepath.Clean(p)
	t, ok := specLookupTables[p]
	if !ok {
		t = SpecLookupTable{
			Table:        map[SpecLocator]string{},
			SpecBasePath: p,
		}
	}
	return t.Lookup(locator)
}

type SpecLookupTable struct {
	Table        map[SpecLocator]string // maps from spec locator to the spec rel-path.
	SpecBasePath string
	isInit       bool
}

func (t *SpecLookupTable) Lookup(locator SpecLocator) (string, error) {
	if !t.isInit {
		if err := t.initOnce(); err != nil {
			return "", err
		}
	}
	path, ok := t.Table[locator]
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
		if err := filepath.Walk(t.SpecBasePath,
			func(p string, info os.FileInfo, err error) error {
				p, _ = filepath.Abs(p)
				t.SpecBasePath, _ = filepath.Abs(t.SpecBasePath)
				relPath := strings.TrimPrefix(p, t.SpecBasePath)
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

				spec, err := swaggerloader.LoadSwagger(t.SpecBasePath + relPath)
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
		t.Table[key.(SpecLocator)] = value.(string)
		return true
	})
	t.isInit = true
	return nil
}
