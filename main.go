package main

import (
	"bosh-blob-experiment/cmd"
	"context"
	"io"
	"log"
	"os"

	cli "github.com/urfave/cli/v3"
)

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
				Action:      cmd.GenerateReport,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "version",
						Aliases:     []string{"v"},
						Usage:       "version regex",
						DefaultText: ".*",
					},
				},
			},
			{
				Name:        "mklocal",
				Aliases:     []string{"make-local"},
				Usage:       "Make this blobstore local",
				Description: "Convert this blobstore to be a local blobstore and download all blobs",
				Action:      cmd.MakeLocal,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "copy",
						Aliases:     []string{"c"},
						Usage:       "copy blobs from existing blobstore",
						DefaultText: "false",
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
