package manifest

// <projectDir>/config/blobs.yml
type Blob struct {
	Size     int
	ObjectId string `yaml:"object_id"`
	Sha      string
}

type Build struct {
	Version     string
	BlobstoreId string `yaml:"blobstore_id"`
	Sha1        string
}

// <projectDir>/.final_builds/**/index.yml
type PackageManifest struct {
	Builds        map[string]Build
	FormatVersion string
}

type Job struct {
	Name        string
	Version     string
	Fingerprint string
	Sha1        string
	Packages    []string
}

type Package struct {
	Name         string
	Version      string
	Fingerprint  string
	Sha1         string
	Dependencies []string
}

// <projectDir>/releases/**/!index.yml
type ReleaseManifest struct {
	Name              string
	Version           string
	CommitHash        string
	UncommitedChanges bool
	Jobs              []Job
	Packages          []Package
}
