package cmd

import (
	"bosh-blob-experiment/blobstore"
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	cli "github.com/urfave/cli/v3"
)

func MakeLocal(_ context.Context, cmd *cli.Command) error {
	projectDir := cmd.String("project")

	newBlobstoreConfig := blobstore.FinalBlobstore{
		Provider: "local",
		Options: map[string]interface{}{
			"blobstore_path": "final_blobs",
		},
	}
	newBlobstore, err := blobstore.NewFromBlobstore(projectDir, newBlobstoreConfig)
	if err != nil {
		return err
	}

	if cmd.Bool("copy") {
		// Make certain the blob directory exists
		blobDir := filepath.Join(projectDir, newBlobstoreConfig.Options["blobstore_path"].(string))
		_, err = os.Stat(blobDir)
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(blobDir, os.ModePerm)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// Get existing configuration
		oldBlobstore, err := blobstore.NewFromConfig(projectDir)
		if err != nil {
			return err
		}

		blobs, err := oldBlobstore.List()
		if err != nil {
			return err
		}
		for _, blob := range blobs {
			log.Printf("copying blob '%s'", blob)

			tmp, err := os.CreateTemp("", "blob-")
			if err != nil {
				return err
			}
			defer func() {
				tmp.Close()
				os.Remove(tmp.Name())
			}()

			err = oldBlobstore.Get(blob, tmp)
			if err != nil {
				return err
			}

			err = newBlobstore.Put(tmp, blob)
			if err != nil {
				return err
			}
		}
	}

	// TODO save blobstore config!
	return nil
}
