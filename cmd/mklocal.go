package cmd

import (
	"bosh-blob-experiment/blobstore"
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	cli "github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

func MakeLocal(_ context.Context, cmd *cli.Command) error {
	projectDir := cmd.String("project")
	blobstorePath := cmd.String("directory")

	if projectDir == "" {
		return errors.New("project directory must be set")
	} else if blobstorePath == "" {
		return errors.New("blobstore subdirectory must be set")
	}

	// Get existing configuration
	oldBlobstore, oldBlobstoreConfig, err := blobstore.NewFromConfig(projectDir)
	if err != nil {
		return err
	}
	if oldBlobstoreConfig.Blobstore.Provider == "local" {
		return errors.New("blobstore is already configured as a local blobstore")
	}

	// Create the new configuration
	newBlobstoreConfig := &blobstore.FinalManifest{
		FinalName: oldBlobstoreConfig.FinalName,
		Blobstore: blobstore.FinalBlobstore{
			Provider: "local",
			Options: map[string]any{
				"blobstore_path": blobstorePath,
			},
		},
	}
	newBlobstore, err := blobstore.NewFromManifest(projectDir, newBlobstoreConfig)
	if err != nil {
		return err
	}

	// Make certain the blob directory exists
	blobDir := filepath.Join(projectDir, blobstorePath)
	_, err = os.Stat(blobDir)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("creating project subdirectory '%s' for blobs", blobstorePath)
		err = os.Mkdir(blobDir, os.ModePerm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if cmd.Bool("copy") {
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

	// Save the new blobstore config
	configDir := filepath.Join(projectDir, "config")
	finalYaml := filepath.Join(configDir, "final.yml")
	finalOld := filepath.Join(configDir, "final.old")
	privateYaml := filepath.Join(configDir, "private.yml")
	privateOld := filepath.Join(configDir, "private.old")

	_, err = os.Stat(finalYaml)
	if err == nil {
		log.Printf("Renaming '%s' to '%s'", finalYaml, finalOld)
		err = os.Rename(finalYaml, finalOld)
	}

	_, err = os.Stat(privateYaml)
	if err == nil {
		log.Printf("Renaming '%s' to '%s'", privateYaml, privateOld)
		err = os.Rename(privateYaml, privateOld)
	}

	log.Printf("Writing new configuration to '%s'", finalYaml)
	bytes, err := yaml.Marshal(newBlobstoreConfig)
	if err != nil {
		return err
	}

	file, err := os.Create(finalYaml)
	if err != nil {
		return err
	}

	_, err = file.Write(bytes)
	return err
}
