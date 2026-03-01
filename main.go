package main

import (
	"bosh-blob-experiment/blobstore"
	"bosh-blob-experiment/manifest"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"maps"
	"os"
	"path/filepath"
	"regexp"

	tbl "github.com/rodaine/table"
	cli "github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

func findProjectReleaseManifests(projectDir string) ([]manifest.ReleaseManifest, error) {
	var releases []manifest.ReleaseManifest
	finalBuildsDir := filepath.Join(projectDir, "releases")
	err := filepath.WalkDir(finalBuildsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if "index.yml" != d.Name() {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel := manifest.ReleaseManifest{}
			err = yaml.Unmarshal(b, &rel)
			if err != nil {
				return err
			}
			releases = append(releases, rel)
		}
		return nil
	})
	return releases, err
}

func findProjectBlobs(projectDir string) (map[string]manifest.Build, error) {
	blobMap := map[string]manifest.Build{}
	finalBuildsDir := filepath.Join(projectDir, ".final_builds")
	err := filepath.WalkDir(finalBuildsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if "index.yml" == d.Name() {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			pkg := manifest.PackageManifest{}
			err = yaml.Unmarshal(b, &pkg)
			if err != nil {
				return err
			}
			maps.Copy(blobMap, pkg.Builds)
		}
		return nil
	})
	return blobMap, err
}

func generateReport(_ context.Context, cmd *cli.Command) error {
	projectDir := cmd.String("project")
	versionRegex := cmd.String("version")
	if versionRegex == "" {
		versionRegex = "*"
	}

	blobs, err := findProjectBlobs(projectDir)
	if err != nil {
		return err
	}
	releases, err := findProjectReleaseManifests(projectDir)
	if err != nil {
		return err
	}
	blobstore, err := blobstore.New(projectDir)
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

func main() {
	log.SetOutput(io.Discard)
	cmd := &cli.Command{
		Name:        "bbx",
		Description: "bosh blob experiment",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "project",
				Aliases:     []string{"p"},
				Usage:       "project directory",
				Sources:     cli.EnvVars("BBX_PROJECT"),
				Local:       false,
				DefaultText: ".",
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Usage:       "enable logging",
				Local:       true,
				DefaultText: "false",
				Action: func(_ context.Context, _ *cli.Command, _ bool) error {
					log.SetOutput(os.Stderr)
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "report",
				Usage:       "Report on blob status",
				Description: "Generate a report of expected blobs for each final release version",
				Action:      generateReport,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "version",
						Aliases: []string{"v"},
						Usage:   "version regex",
					},
				},
			},
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		panic(err)
	}

	// blobMap := map[string]string{}
	// finalBuildsDir := filepath.Join(projectDir, ".final_builds")
	// err := filepath.WalkDir(finalBuildsDir, func(path string, d fs.DirEntry, err error) error {
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if d.IsDir() {
	// 		return nil
	// 	}
	// 	if "index.yml" == d.Name() {
	// 		b, err := os.ReadFile(path)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		pkg := manifest.PackageManifest{}
	// 		err = yaml.Unmarshal(b, &pkg)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		for _, b := range pkg.Builds {
	// 			blobMap[b.Sha1] = b.BlobstoreId
	// 		}
	// 	}
	// 	return nil
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(blobMap)

	// for i, arg := range flag.Args() {
	// 	fmt.Printf("%d: %s\n", i, arg)

	// 	tmp, err := os.CreateTemp("", "release-")
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	defer func() {
	// 		tmp.Close()
	// 		os.Remove(tmp.Name())
	// 	}()
	// 	fmt.Printf("tmp=%s\n", tmp.Name())

	// 	resp, err := http.Get(arg)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	defer resp.Body.Close()

	// 	n, err := io.Copy(tmp, resp.Body)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Printf("Bytes copied: %d\n", n)

	// 	_, err = tmp.Seek(0, io.SeekStart)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	gz, err := gzip.NewReader(tmp)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	tarFile := tar.NewReader(gz)
	// 	for {
	// 		h, err := tarFile.Next()
	// 		if err == io.EOF {
	// 			break
	// 		} else if err != nil {
	// 			panic(err)
	// 		}

	// 		if "release.MF" == h.Name {
	// 			meta := manifest.ReleaseManifest{}
	// 			decoder := yaml.NewDecoder(tarFile)
	// 			err = decoder.Decode(&meta)
	// 			if err != nil {
	// 				panic(err)
	// 			}
	// 			fmt.Println(meta)
	// 		}
	// 	}
	// }
}
