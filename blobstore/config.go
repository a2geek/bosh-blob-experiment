package blobstore

type FinalBlobstore struct {
	Provider string
	Options  map[string]any
}

type FinalManifest struct {
	Blobstore FinalBlobstore
	FinalName string `yaml:"final_name"`
}
