package util

import (
	"bosh-blob-experiment/manifest"
	"io/fs"
	"maps"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func FindProjectBlobs(projectDir string) (map[string]manifest.Build, error) {
	blobMap := map[string]manifest.Build{}
	finalBuildsDir := filepath.Join(projectDir, ".final_builds")
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
			pkg := manifest.PackageManifest{}
			err = yaml.Unmarshal(b, &pkg)
			if err != nil {
				return err
			}
			maps.Copy(blobMap, pkg.Builds)
		}
		return nil
	})
	return blobMap, err
}
