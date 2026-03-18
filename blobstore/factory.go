package blobstore

import (
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Blobstore interface {
	Get(src string, dest io.WriterAt) error
	Put(src io.ReadSeeker, dest string) error
	Exists(name string) (bool, error)
	List() ([]string, error)
}

func NewFromConfig(projectDir string) (Blobstore, *FinalManifest, error) {
	finalYaml := filepath.Join(projectDir, "config", "final.yml")

	b, err := os.ReadFile(finalYaml)
	if err != nil {
		return nil, nil, err
	}
	manifest := &FinalManifest{}
	err = yaml.Unmarshal(b, &manifest)
	if err != nil {
		return nil, nil, err
	}

	privateYaml := filepath.Join(projectDir, "config", "private.yml")
	b, err = os.ReadFile(privateYaml)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, nil, err
		}
	} else {
		// The 'private.yml' file appears to be mostly identical to the main manifest, so merge the options section
		privateManifest := FinalManifest{}
		err = yaml.Unmarshal(b, &privateManifest)
		if err != nil {
			return nil, nil, err
		}
		maps.Copy(manifest.Blobstore.Options, privateManifest.Blobstore.Options)
	}
	blobstore, err := NewFromManifest(projectDir, manifest)
	return blobstore, manifest, err
}

func NewFromManifest(projectDir string, config *FinalManifest) (Blobstore, error) {
	log.Printf("blobstore type '%s' found", config.Blobstore.Provider)
	switch config.Blobstore.Provider {
	case "s3", "gcs":
		return NewS3Blobstore(config.Blobstore)
	case "local":
		blobstoreDir := filepath.Join(projectDir, config.Blobstore.Options["blobstore_path"].(string))
		return NewLocalBlobstore(blobstoreDir), nil
	default:
		return nil, fmt.Errorf("blobstore of type '%s' not supported", config.Blobstore.Provider)
	}
}
