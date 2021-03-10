package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/magodo/azure-rest-api-cov/analyzer"

	log "github.com/sirupsen/logrus"

	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	"github.com/pkg/profile"
	"github.com/urfave/cli/v2"
)

func main() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}

	app := cli.App{
		Name:  "azure-rest-api-cov",
		Usage: "Azure rest API model property coverage",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "workspace-dir",
				Aliases: []string{"w"},
				Usage:   "The path to the workspace for caching the azure-rest-api repository and other stuffs",
				Value:   filepath.Join(cacheDir, "azure-rest-api-cov"),
			},
			&cli.BoolFlag{
				Name:    "workspace-force-new",
				Aliases: []string{"f"},
				Usage:   "Whether to force recreate the workspace",
			},
			&cli.BoolFlag{
				Name:    "index-renew",
				Aliases: []string{"r"},
				Usage:   "Whether to force regenerate the repository index in the workspace",
			},
			&cli.StringFlag{
				Name:  "spec-ref",
				Usage: "The azure-rest-api-spec repo ref object to checkout",
			},
			&cli.StringFlag{
				Name:    "github-username",
				Aliases: []string{"u"},
				Usage:   "github username that might be needed for accessing the grant repo",
			},
			&cli.StringFlag{
				Name:    "github-token",
				Aliases: []string{"t"},
				Usage:   "github token that might be needed for accessing the grant repo",
			},
			&cli.StringFlag{
				Name:  "log-level",
				Value: "info",
				Usage: `Log level: can be one of "debug", "info", "warn", "fatal"`,
			},

			// profiles
			&cli.BoolFlag{
				Name:  "profile-cpu",
				Usage: "Profile CPU",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "report",
				Usage: "Generate coverage report",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "input",
						Required: true,
						Aliases:  []string{"i"},
						Usage:    "The path to the schema input file, which is expected to hold a JSON array of input schemas",
					},
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a"},
						Usage:   "Whether to export the RP level coverage",
					},
					&cli.BoolFlag{
						Name:    "predict",
						Aliases: []string{"p"},
						Usage:   "Whether to predict coverage for newer API versions",
					},
					&cli.StringFlag{
						Name:    "grant-repo-url",
						Aliases: []string{"g"},
						Usage:   "The grant repo URL",
					},
					&cli.StringFlag{
						Name:  "grant-ref",
						Usage: "The grant repo ref object to use",
					},
					&cli.StringFlag{
						Name:    "additional-message",
						Aliases: []string{"m"},
						Usage:   "Additional message that will include in the coverage report",
					},
					&cli.BoolFlag{
						Name:  "migrate-grant-hint",
						Usage: `Report possible grant migrations (e.g. there exists some grant for a model in old API version, but absent in the current version) in a file "possible_grant_migration.txt"`,
					},
					&cli.BoolFlag{
						Name:  "try-migrate",
						Usage: `Try to migrate best match grants from older/newer API version and use it to grant the model of current API version`,
					},
					&cli.StringSliceFlag{
						Name:  "ignore-schema",
						Usage: `Ignore the schemas that matches the specified pattern from calculating the property coverage. The pattern is in the format of: <API Path>:<Operation>:<API Version>`,
					},
				},
				Action: func(ctx *cli.Context) error {
					initLog(ctx.String("log-level"))

					if ctx.Bool("profile-cpu") {
						defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
					}

					workspaceDir := ctx.String("workspace-dir")
					force := ctx.Bool("workspace-force-new")
					regenIndex := ctx.Bool("index-renew")
					specRepoRef := ctx.String("spec-ref")
					githubUsername := ctx.String("github-username")
					githubToken := ctx.String("github-token")
					inputPath := ctx.String("input")

					all := ctx.Bool("all")
					predict := ctx.Bool("predict")
					grantRepoUrl := ctx.String("grant-repo-url")
					grantRepoRef := ctx.String("grant-ref")
					additionalMessage := ctx.String("additional-message")
					migrateGrantHint := ctx.Bool("migrate-grant-hint")
					tryMigrate := ctx.Bool("try-migrate")
					ignoreSchema := ctx.StringSlice("ignore-schema")

					if grantRepoRef != "" && grantRepoUrl == "" {
						return fmt.Errorf("`-grant-ref` has to be used together with `-g`")
					}

					var auth transport.AuthMethod
					if githubToken != "" && githubUsername != "" {
						auth = &http.BasicAuth{
							Username: githubUsername,
							Password: githubToken,
						}
					}

					wspOpt := workspace.InitOption{
						WorkspaceDir: workspaceDir,
						Force:        force,
						RegenIndex:   regenIndex,
						RepoAuth:     auth,
						SpecRepoRef:  specRepoRef,
						GrantRepoURL: grantRepoUrl,
						GrantRepoRef: grantRepoRef,
					}

					schemaInputFile, err := os.Open(inputPath)
					if err != nil {
						return err
					}
					defer schemaInputFile.Close()

					runCtx := analyzer.RunContext{
						WorkspaceOpt:      wspOpt,
						Predict:           predict,
						All:               all,
						AdditionalMessage: additionalMessage,
						MigrateGrantHint:  migrateGrantHint,
						TryMigrate:        tryMigrate,
						IgnoreSchema:      ignoreSchema,
					}

					output, err := analyzer.Run(schemaInputFile, runCtx)
					if err != nil {
						return err
					}
					fmt.Println(output)
					return nil
				},
			},
			{
				Name:  "migrate-grant",
				Usage: "Migrate grants",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "grant-repo-dir",
						Aliases:  []string{"g"},
						Usage:    "The grant repo root dir",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "rp",
						Usage:    "The RP that the grants belong to to migrate",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "spec",
						Usage: "The spec that the grants belong to to migrate",
					},
					&cli.StringFlag{
						Name:     "src-version",
						Usage:    "The source API version (the version appears in spec file path) to migrate from",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "dst-version",
						Usage:    "The destination API version (the version appears in spec file path) to migrate to",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "The output directory for the migrated grant files",
					},
				},
				Action: func(ctx *cli.Context) error {
					initLog(ctx.String("log-level"))

					if ctx.Bool("profile-cpu") {
						defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
					}

					workspaceDir := ctx.String("workspace-dir")
					force := ctx.Bool("workspace-force-new")
					regenIndex := ctx.Bool("index-renew")
					specRepoRef := ctx.String("spec-ref")
					githubUsername := ctx.String("github-username")
					githubToken := ctx.String("github-token")

					grantRepoDir := ctx.String("grant-repo-dir")
					rp := ctx.String("rp")
					spec := ctx.String("spec")
					srcVersion := ctx.String("src-version")
					dstVersion := ctx.String("dst-version")
					outDir := ctx.String("output")

					var auth transport.AuthMethod
					if githubToken != "" && githubUsername != "" {
						auth = &http.BasicAuth{
							Username: githubUsername,
							Password: githubToken,
						}
					}

					wspOpt := workspace.InitOption{
						WorkspaceDir: workspaceDir,
						Force:        force,
						RegenIndex:   regenIndex,
						RepoAuth:     auth,
						SpecRepoRef:  specRepoRef,
					}

					return analyzer.MigrateGrant(wspOpt, grantRepoDir, rp, spec, srcVersion, dstVersion, outDir)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func initLog(logLevel string) {
	// Initialize logger
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	switch logLevel {
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
}
