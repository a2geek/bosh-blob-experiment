package manifest

import (
	"maps"
	"regexp"
)

type Model struct {
	projectDir       string
	blobs            map[string]Blob
	packageManifests []PackageManifest
	releaseManifests []ReleaseManifest
}

func (m *Model) FindReleases(pattern string) ([]ReleaseManifest, error) {
	var matches []ReleaseManifest
	for _, release := range m.releaseManifests {
		match, err := regexp.MatchString(pattern, release.Version)
		if err != nil {
			return nil, err
		}
		if !match {
			continue
		}
		matches = append(matches, release)
	}
	return matches, nil
}

func (m *Model) FindAllBuilds() map[string]Build {
	blobs := map[string]Build{}
	for _, pkg := range m.packageManifests {
		maps.Copy(blobs, pkg.Builds)
	}
	return blobs
}

func (m *Model) FindBuildByVersion(version string) *Build {
	for _, pkg := range m.packageManifests {
		for _, build := range pkg.Builds {
			if build.Version == version {
				return &build
			}
		}
	}
	return nil
}

func (m *Model) ConfigBlobs() map[string]Blob {
	blobs := map[string]Blob{}
	maps.Copy(blobs, m.blobs)
	return blobs
}
