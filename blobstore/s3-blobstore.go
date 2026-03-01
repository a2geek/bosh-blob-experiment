package blobstore

import (
	"bytes"
	"encoding/json"

	"github.com/cloudfoundry/bosh-s3cli/client"
	"github.com/cloudfoundry/bosh-s3cli/config"
)

func NewS3Blobstore(manifest finalBlobstore) (Blobstore, error) {
	if manifest.Provider == "s3" {
		if _, ok := manifest.Options["region"]; !ok {
			manifest.Options["region"] = "us-east-1"
		}
	}

	b, err := json.Marshal(manifest.Options)
	if err != nil {
		return nil, err
	}

	s3Config, err := config.NewFromReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	s3Client, err := client.NewAwsS3Client(&s3Config)
	if err != nil {
		return nil, err
	}

	return client.New(s3Client, &s3Config), nil
}
