package cmd

import (
	"bosh-blob-experiment/blobstore"
	"bosh-blob-experiment/util"
	"context"
	"fmt"
	"regexp"

	tbl "github.com/rodaine/table"
	cli "github.com/urfave/cli/v3"
)

func GenerateReport(_ context.Context, cmd *cli.Command) error {
	projectDir := cmd.String("project")
	versionRegex := cmd.String("version")

	blobs, err := util.FindProjectBlobs(projectDir)
	if err != nil {
		return err
	}
	releases, err := util.FindProjectReleaseManifests(projectDir)
	if err != nil {
		return err
	}
	blobstore, err := blobstore.NewFromConfig(projectDir)
	if err != nil {
		return err
	}

	count := 0
	var versions []string
	for _, release := range releases {
		match, err := regexp.MatchString(versionRegex, release.Version)
		if err != nil {
			return err
		}
		if !match {
			continue
		}
		count += len(release.Jobs)
		count += len(release.Packages)
		versions = append(versions, release.Version)
	}
	if count > 50 {
		return fmt.Errorf("too many blobs to lookup; found %d but max is 50; versions found %v", count, versions)
	}

	tbl := tbl.New("Version", "Type", "Name", "Blob Name", "Present?")
	for _, release := range releases {
		match, err := regexp.MatchString(versionRegex, release.Version)
		if err != nil {
			return err
		}
		if !match {
			continue
		}
		version := release.Version // Note: Only showing on the first entry
		for _, job := range release.Jobs {
			blobId := "-"
			present := false
			blob, ok := blobs[job.Version]
			if ok {
				blobId = blob.BlobstoreId
				present, err = blobstore.Exists(blobId)
				if err != nil {
					return err
				}
			}
			tbl.AddRow(version, "Job", job.Name, blobId, present)
			version = ""
		}
		for _, pkg := range release.Packages {
			blobId := "-"
			present := false
			blob, ok := blobs[pkg.Version]
			if ok {
				blobId = blob.BlobstoreId
				present, err = blobstore.Exists(blobId)
				if err != nil {
					return err
				}
			}
			tbl.AddRow(version, "Package", pkg.Name, blobId, present)
			version = ""
		}
	}
	tbl.Print()
	return nil
}
