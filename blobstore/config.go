package blobstore

type finalBlobstore struct {
	Provider string
	Options  map[string]interface{}
}

type finalManifest struct {
	Blobstore finalBlobstore
	FinalName string `yaml:"final_name"`
}
