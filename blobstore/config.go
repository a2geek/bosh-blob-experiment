package blobstore

type FinalBlobstore struct {
	Provider string
	Options  map[string]interface{}
}

type FinalManifest struct {
	Blobstore FinalBlobstore
	FinalName string `yaml:"final_name"`
}
