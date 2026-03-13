package manifest

import (
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Load(projectDir string) (*Model, error) {
	blobYaml := filepath.Join(projectDir, "config", "blobs.yml")
	finalBuildsDir := filepath.Join(projectDir, ".final_builds")
	releasesDir := filepath.Join(projectDir, "releases")

	packageManifests := []PackageManifest{}
	err := filepath.WalkDir(finalBuildsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if "index.yml" == d.Name() {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			pkg := PackageManifest{}
			err = yaml.Unmarshal(b, &pkg)
			if err != nil {
				return err
			}
			packageManifests = append(packageManifests, pkg)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var releases []ReleaseManifest
	err = filepath.WalkDir(releasesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if "index.yml" != d.Name() {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel := ReleaseManifest{}
			err = yaml.Unmarshal(b, &rel)
			if err != nil {
				return err
			}
			releases = append(releases, rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	blobs := map[string]Blob{}
	b, err := os.ReadFile(blobYaml)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(b, &blobs)
	if err != nil {
		return nil, err
	}

	return &Model{
		projectDir:       projectDir,
		blobs:            blobs,
		packageManifests: packageManifests,
		releaseManifests: releases,
	}, nil
}
