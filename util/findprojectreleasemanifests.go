package util

import (
	"bosh-blob-experiment/manifest"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func FindProjectReleaseManifests(projectDir string) ([]manifest.ReleaseManifest, error) {
	var releases []manifest.ReleaseManifest
	finalBuildsDir := filepath.Join(projectDir, "releases")
	err := filepath.WalkDir(finalBuildsDir, func(path string, d fs.DirEntry, err error) error {
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
			rel := manifest.ReleaseManifest{}
			err = yaml.Unmarshal(b, &rel)
			if err != nil {
				return err
			}
			releases = append(releases, rel)
		}
		return nil
	})
	return releases, err
}
