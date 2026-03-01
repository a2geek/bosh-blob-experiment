package blobstore

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Blobstore interface {
	Exists(Name string) (bool, error)
}

func New(projectDir string) (Blobstore, error) {
	finalYaml := filepath.Join(projectDir, "config", "final.yml")

	b, err := os.ReadFile(finalYaml)
	if err != nil {
		return nil, err
	}
	manifest := finalManifest{}
	err = yaml.Unmarshal(b, &manifest)
	if err != nil {
		return nil, err
	}

	privateYaml := filepath.Join(projectDir, "config", "private.yml")
	b, err = os.ReadFile(privateYaml)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	} else {
		// The 'private.yml' file appears to be mostly identical to the main manifest, so merge the options section
		privateManifest := finalManifest{}
		err = yaml.Unmarshal(b, &privateManifest)
		if err != nil {
			return nil, err
		}
		maps.Copy(manifest.Blobstore.Options, privateManifest.Blobstore.Options)
	}

	switch manifest.Blobstore.Provider {
	case "s3", "gcs":
		return NewS3Blobstore(manifest.Blobstore)
	default:
		return nil, fmt.Errorf("blobstore of type '%s' not supported", manifest.Blobstore.Provider)
	}
}
