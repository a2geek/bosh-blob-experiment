package cmd

import (
	"bosh-blob-experiment/blobstore"
	"context"
	"path/filepath"

	cli "github.com/urfave/cli/v3"
)

func CopyBlobs(_ context.Context, cmd *cli.Command) error {
	projectDir := cmd.String("project")
	oldProjectDir := ""
	forceFlag := cmd.Bool("force")
	finalFile := cmd.String("final")
	privateFile := cmd.String("private")

	if finalFile == "" && privateFile == "" {
		finalFile = filepath.Join(projectDir, "config", "final.old")
		privateFile = filepath.Join(projectDir, "config", "private.old")
		oldProjectDir = projectDir
	}

	currentBlobstore, _, err := blobstore.NewFromConfig(projectDir)
	if err != nil {
		return err
	}

	oldBlobstore, _, err := blobstore.NewFromFiles(oldProjectDir, finalFile, privateFile, false)
	if err != nil {
		return err
	}

	return blobstore.Copy(oldBlobstore, currentBlobstore, forceFlag)
}
