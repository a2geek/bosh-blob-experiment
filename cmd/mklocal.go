package cmd

import (
	"bosh-blob-experiment/blobstore"
	"context"
	"os"

	cli "github.com/urfave/cli/v3"
)

func MakeLocal(_ context.Context, cmd *cli.Command) error {
	projectDir := cmd.String("project")

	newBlobstore, err := blobstore.NewFromBlobstore(blobstore.FinalBlobstore{
		Provider: "local",
		Options: map[string]interface{}{
			"blobstore_path": "final_blobs",
		},
	})
	if err != nil {
		return err
	}

	if cmd.Bool("copy") {
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
