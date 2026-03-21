package cmd

import (
	"bosh-blob-experiment/blobstore"
	"bosh-blob-experiment/manifest"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"

	tbl "github.com/rodaine/table"
	cli "github.com/urfave/cli/v3"
)

func GenerateReport(_ context.Context, cmd *cli.Command) error {
	projectDir := cmd.String("project")
	versionRegex := cmd.String("version")
	verifyFlag := cmd.Bool("verify")
	allFlag := cmd.Bool("all")
	maxBlobs := cmd.Int("max-blobs")

	model, err := manifest.Load(projectDir)
	if err != nil {
		return nil
	}

	blobstore, _, err := blobstore.NewFromConfig(projectDir)
	if err != nil {
		return err
	}

	releases, err := model.FindReleases(versionRegex)
	if err != nil {
		return err
	}

	var blobList []string
	if allFlag {
		blobList, err = blobstore.List()
		if err != nil {
			return err
		}
	}

	count := 0
	var versions []string
	for _, release := range releases {
		count += len(release.Jobs)
		count += len(release.Packages)
		versions = append(versions, release.Version)
	}
	if count > maxBlobs {
		return fmt.Errorf("too many blobs to lookup; found %d but max is %d; versions found %v", count, maxBlobs, versions)
	}

	blobs := model.FindAllBuilds()

	headers := []any{"Version", "Name", "Blob Name", "Present?"}
	if verifyFlag {
		headers = append(headers, "SHA?")
	}
	tbl := tbl.New(headers...)
	for _, release := range releases {
		version := release.Version // Note: Only showing on the first entry
		for _, job := range release.Jobs {
			blobId, err := addRow(tbl, blobstore, blobs, job.Version, version, fmt.Sprintf("job/%s", job.Name), verifyFlag, job.Sha1)
			if err != nil {
				return err
			}
			blobList = removeFromList(blobList, blobId)
			version = ""
		}
		for _, pkg := range release.Packages {
			blobId, err := addRow(tbl, blobstore, blobs, pkg.Version, version, fmt.Sprintf("package/%s", pkg.Name), verifyFlag, pkg.Sha1)
			if err != nil {
				return err
			}
			blobList = removeFromList(blobList, blobId)
			version = ""
		}
	}
	for name, blob := range model.ConfigBlobs() {
		fakeBlobMap := map[string]manifest.Build{
			name: {
				Version:     name,
				BlobstoreId: blob.ObjectId,
				Sha1:        blob.Sha,
			},
		}
		blobId, err := addRow(tbl, blobstore, fakeBlobMap, name, "-", name, verifyFlag, blob.Sha)
		if err != nil {
			return err
		}
		blobList = removeFromList(blobList, blobId)
	}
	for _, blobId := range blobList {
		fakeBlobMap := map[string]manifest.Build{
			blobId: {
				Version:     blobId,
				BlobstoreId: blobId,
				Sha1:        "",
			},
		}
		addRow(tbl, blobstore, fakeBlobMap, blobId, "-", "<unknown>", verifyFlag, "")
	}
	tbl.Print()
	return nil
}

func addRow(tbl tbl.Table, blobstore blobstore.Blobstore, blobs map[string]manifest.Build, blobVersion, version, name string, verifyFlag bool, expectedSha string) (string, error) {
	blobId := "-"
	present := false
	blob, ok := blobs[blobVersion]
	var err error
	if ok {
		blobId = blob.BlobstoreId
		present, err = blobstore.Exists(blobId)
		if err != nil {
			return "", err
		}
	}
	row := []any{version, name, blobId, present}
	if verifyFlag {
		if !present {
			row = append(row, "does NOT match")
		} else if expectedSha == "" {
			row = append(row, "<unknown>")
		} else {
			tmp, err := os.CreateTemp("", "blob-")
			if err != nil {
				return "", err
			}
			defer func() {
				tmp.Close()
				os.Remove(tmp.Name())
			}()

			err = blobstore.Get(blobId, tmp)
			if err != nil {
				return "", err
			}

			sha := sha1.New()
			prefix := ""
			if strings.HasPrefix(expectedSha, "sha256:") {
				sha = sha256.New()
				prefix = "sha256:"
			}

			_, err = io.Copy(sha, tmp)
			if err != nil {
				return "", err
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
	return blobId, nil
}

func removeFromList(slice []string, target string) []string {
	newSlice := []string{}
	for _, str := range slice {
		if str != target {
			newSlice = append(newSlice, str)
		}
	}
	return newSlice
}
