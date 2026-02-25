package manifest

type Build struct {
	Version     string
	BlobstoreId string `yaml:"blobstore_id"`
	Sha1        string
}

type PackageManifest struct {
	Builds        map[string]Build
	FormatVersion string
}
