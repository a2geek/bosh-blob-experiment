package blobstore

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cloudfoundry/bosh-s3cli/client"
	"github.com/cloudfoundry/bosh-s3cli/config"
)

func NewS3Blobstore(manifest FinalBlobstore) (Blobstore, error) {
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

	blobstore := s3Blobstore{
		s3Config:           s3Config,
		s3Client:           s3Client,
		s3CompatibleClient: client.New(s3Client, &s3Config),
	}
	return &blobstore, nil
}

type s3Blobstore struct {
	s3Config           config.S3Cli
	s3Client           *s3.Client
	s3CompatibleClient client.S3CompatibleClient
}

func (c *s3Blobstore) Get(src string, dest io.WriterAt) error {
	return c.s3CompatibleClient.Get(src, dest)
}

func (c *s3Blobstore) Put(src io.ReadSeeker, dest string) error {
	return c.s3CompatibleClient.Put(src, dest)
}

func (c *s3Blobstore) Exists(dest string) (bool, error) {
	return c.s3CompatibleClient.Exists(dest)
}

func (c *s3Blobstore) List() ([]string, error) {
	output, err := c.s3Client.ListObjects(context.Background(), &s3.ListObjectsInput{
		Bucket: &c.s3Config.BucketName,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("list objects for bucket '%s' returned %d objects", c.s3Config.BucketName, len(output.Contents))

	var list []string
	for _, object := range output.Contents {
		list = append(list, *object.Key)
	}
	return list, nil
}
