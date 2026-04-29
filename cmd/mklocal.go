package cmd

import (
	"bosh-blob-experiment/blobstore"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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

	// Make certain the blob directories exist (/blob/ and /<blobstorePath>/)
	err = ensureDirectoriesExists(projectDir, "blobs", blobstorePath)
	if err != nil {
		return err
	}

	if cmd.Bool("copy") {
		newBlobstore, err := blobstore.NewFromManifest(projectDir, newBlobstoreConfig)
		if err != nil {
			return err
		}

		err = blobstore.Copy(oldBlobstore, newBlobstore, false)
		if err != nil {
			return err
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

	// Update .gitattributes file
	gitAttributesFile := filepath.Join(projectDir, ".gitattributes")
	gitAttributes := ""

	_, err = os.Stat(gitAttributesFile)
	if err == nil {
		log.Println("reading existing .gitattributes file")
		b, err := os.ReadFile(gitAttributesFile)
		if err != nil {
			return err
		}
		gitAttributes = string(b)
	}

	updatedGitAttributes := checkGitAttributes(gitAttributes, "blobs", blobstorePath)
	if gitAttributes != updatedGitAttributes {
		err = os.WriteFile(gitAttributesFile, []byte(updatedGitAttributes), os.ModePerm)
		if err != nil {
			return err
		}
	}

	return err
}

func ensureDirectoriesExists(baseDir string, subdirectories ...string) error {
	for _, subdir := range subdirectories {
		fullDir := filepath.Join(baseDir, subdir)
		_, err := os.Stat(fullDir)
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("creating project subdirectory '%s' for blobs", subdir)
			err = os.Mkdir(fullDir, os.ModePerm)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

func checkGitAttributes(gitAttributes string, paths ...string) string {
	for _, path := range paths {
		if !strings.Contains(gitAttributes, path) {
			log.Printf("adding entry to .gitattributes file for '/%s/'", path)
			if gitAttributes != "" && !strings.HasSuffix(gitAttributes, "\n") {
				gitAttributes += "\n"
			}
			gitAttributes += fmt.Sprintf("/%s/** filter=lfs diff=lfs merge=lfs -text\n", path)
		} else {
			log.Printf(".gitattributes file already has entry for '%s'", path)
		}
	}
	return gitAttributes
}
