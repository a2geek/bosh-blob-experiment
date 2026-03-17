package main

import (
	"bosh-blob-experiment/cmd"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	cli "github.com/urfave/cli/v3"
)

func main() {
	debugFlag := false
	log.SetOutput(io.Discard)
	cmd := &cli.Command{
		Name:        "bbx",
		Description: "bosh blob experiment",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "project",
				Aliases:     []string{"p"},
				Usage:       "project directory",
				Sources:     cli.EnvVars("BBX_PROJECT"),
				DefaultText: ".",
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Usage:       "enable logging",
				Local:       true,
				DefaultText: "false",
				Action: func(_ context.Context, _ *cli.Command, _ bool) error {
					log.SetOutput(os.Stderr)
					return nil
				},
			},
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "dump stacktrace for any error raised",
				Local:       true,
				DefaultText: "false",
				Action: func(_ context.Context, _ *cli.Command, b bool) error {
					debugFlag = b
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "report",
				Aliases:     []string{"blobs-report"},
				Usage:       "Report on blob status",
				Description: "Generate a report of expected blobs for each final release version",
				Action:      cmd.GenerateReport,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "version",
						Aliases:     []string{"v"},
						Usage:       "version regex",
						DefaultText: ".*",
					},
					&cli.BoolFlag{
						Name:    "verify",
						Aliases: []string{"V"},
						Usage:   "Verify the SHA against the file. WARNING: This may take a bit when blobstore is remote.",
					},
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a"},
						Usage:   "Include all blobs found in the blobstore.",
					},
				},
			},
			{
				Name:        "mklocal",
				Aliases:     []string{"make-blobstore-local"},
				Usage:       "Make this blobstore local",
				Description: "Convert this blobstore to be a local blobstore and download all blobs",
				Action:      cmd.MakeLocal,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "copy",
						Aliases:     []string{"c"},
						Usage:       "copy blobs from existing blobstore",
						DefaultText: "false",
					},
				},
			},
			{
				Name:        "import",
				Aliases:     []string{"import-release-blobs"},
				Usage:       "Import blobs from a release",
				Description: "Fetch a release (file or URL allowed) and import any blobs that aren't present",
				Action:      cmd.ImportRelease,
			},
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		if debugFlag {
			panic(err)
		} else {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
}
