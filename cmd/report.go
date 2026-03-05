package cmd

import (
	"bosh-blob-experiment/blobstore"
	"bosh-blob-experiment/manifest"
	"bosh-blob-experiment/util"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	tbl "github.com/rodaine/table"
	cli "github.com/urfave/cli/v3"
)

func GenerateReport(_ context.Context, cmd *cli.Command) error {
	projectDir := cmd.String("project")
	versionRegex := cmd.String("version")
	verifyFlag := cmd.Bool("verify")

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

	headers := []any{"Version", "Type", "Name", "Blob Name", "Present?"}
	if verifyFlag {
		headers = append(headers, "SHA?")
	}
	tbl := tbl.New(headers...)
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
			addRow(tbl, blobstore, blobs, job.Version, "Job", version, job.Name, verifyFlag, job.Sha1)
			version = ""
		}
		for _, pkg := range release.Packages {
			addRow(tbl, blobstore, blobs, pkg.Version, "Job", version, pkg.Name, verifyFlag, pkg.Sha1)
			version = ""
		}
	}
	tbl.Print()
	return nil
}

func addRow(tbl tbl.Table, blobstore blobstore.Blobstore, blobs map[string]manifest.Build, blobVersion, label, version, name string, verifyFlag bool, expectedSha string) error {
	blobId := "-"
	present := false
	blob, ok := blobs[blobVersion]
	var err error
	if ok {
		blobId = blob.BlobstoreId
		present, err = blobstore.Exists(blobId)
		if err != nil {
			return err
		}
	}
	row := []any{version, label, name, blobId, present}
	if verifyFlag {
		if !present {
			row = append(row, "does NOT match")
		} else {
			tmp, err := os.CreateTemp("", "blob-")
			if err != nil {
				return err
			}
			defer func() {
				tmp.Close()
				os.Remove(tmp.Name())
			}()

			err = blobstore.Get(blobId, tmp)
			if err != nil {
				return err
			}

			sha := sha1.New()
			prefix := ""
			if strings.HasPrefix(expectedSha, "sha256:") {
				sha = sha256.New()
				prefix = "sha256:"
			}

			_, err = io.Copy(sha, tmp)
			if err != nil {
				return err
			}

			result := sha.Sum(nil)

			if expectedSha == fmt.Sprintf("%s%x", prefix, result) {
				row = append(row, "matches")
			} else {
				row = append(row, "does NOT match")
			}
		}
	}
	tbl.AddRow(row...)
	return nil
}
