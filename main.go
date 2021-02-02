package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/magodo/azure-rest-api-cov/analyzer"

	log "github.com/sirupsen/logrus"

	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
)

const usage = `azure-rest-api-cov -i <schema input> <option>`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n", usage)
		flag.PrintDefaults()
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}
	// Required options
	inputPath := flag.String("i", "",
		"The path to the schema input file, which is expected to hold a JSON array of input schemas")

	// Workspace related options
	workspaceDir := flag.String("w", filepath.Join(cacheDir, "azure-rest-api-cov"),
		"The path to the workspace for caching the azure-rest-api repository and other stuffs")
	force := flag.Bool("f", false, "whether to force create the workspace")
	regenIndex := flag.Bool("r", false, "whether to force regenerate the repository index in the workspace")
	specRepoRef := flag.String("spec-ref", "", "The azure-rest-api-spec repo ref object to checkout")

	// Behavior options
	all := flag.Bool("a", false, "whether to export the RP level coverage")
	predict := flag.Bool("p", false, "whether to predict coverage for newer API versions")

	// Grant options
	grantRepoUrl := flag.String("g", "", "the repo URL based on which this tool will honor the grants")
	grantRepoRef := flag.String("grant-ref", "", "The grant repo ref object to checkout")
	repoUser := flag.String("u", "", "github user that might be needed for accessing the grant repo")
	repoToken := flag.String("t", "", "github access token that might be needed for accessing the grant repo")

	// Other options
	logLevel := flag.String("log-level", "info", `Log level: can be one of "debug", "info", "warn", "fatal"`)
	additionalMessage := flag.String("m", "", "Additional message that will include in the coverage report")

	// Parse flag
	flag.Parse()

	// Initialize logger
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	switch *logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	default:
		log.Fatal("unknown log level")
	}

	if *inputPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *grantRepoRef != "" && *grantRepoUrl == "" {
		log.Fatal("`-grant-commit` has to be used together with `-g`")
	}

	var auth transport.AuthMethod

	if *repoToken != "" && *repoUser != "" {
		auth = &http.BasicAuth{
			Username: *repoUser,
			Password: *repoToken,
		}
	}

	wspOpt := workspace.InitOption{
		WorkspaceDir: *workspaceDir,
		Force:        *force,
		RegenIndex:   *regenIndex,
		RepoAuth:     auth,
		SpecRepoRef:  *specRepoRef,
		GrantRepoURL: *grantRepoUrl,
		GrantRepoRef: *grantRepoRef,
	}

	schemaInputFile, err := os.Open(*inputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer schemaInputFile.Close()

	output, err := analyzer.Run(schemaInputFile, wspOpt, *predict, *all, *additionalMessage)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output)
}
