package blobstore

type FinalBlobstore struct {
	Provider string
	Options  map[string]any
}

// FinalManifest is the root representation for 'config/final.yml'
type FinalManifest struct {
	Blobstore FinalBlobstore
	FinalName string `yaml:"final_name"`
}
