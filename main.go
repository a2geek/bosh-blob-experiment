package main

import (
	"bosh-blob-experiment/manifest"
	"context"
	"io/fs"
	"maps"
	"os"
	"path/filepath"

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
	blobs, err := findProjectBlobs(projectDir)
	if err != nil {
		return err
	}
	releases, err := findProjectReleaseManifests(projectDir)
	if err != nil {
		return err
	}
	tbl := tbl.New("Version", "Type", "Name", "Blob Name", "Present?")
	for _, release := range releases {
		version := release.Version // Note: Only showing on the first entry
		for _, job := range release.Jobs {
			blobId := "-"
			present := "No"
			blob, ok := blobs[job.Version]
			if ok {
				blobId = blob.BlobstoreId
				present = "Yes"
			}
			tbl.AddRow(version, "Job", job.Name, blobId, present)
			version = ""
		}
		for _, pkg := range release.Packages {
			blobId := "-"
			present := "No"
			blob, ok := blobs[pkg.Version]
			if ok {
				blobId = blob.BlobstoreId
				present = "Yes"
			}
			tbl.AddRow(version, "Package", pkg.Name, blobId, present)
			version = ""
		}
	}
	tbl.Print()
	return nil
}

func main() {
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
		},
		Commands: []*cli.Command{
			{
				Name:        "report",
				Usage:       "Report on blob status",
				Description: "Generate a report of expected blobs for each final release version",
				Action:      generateReport,
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
