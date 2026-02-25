package main

import (
	"archive/tar"
	"bosh-blob-experiment/manifest"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func main() {
	projectDir := flag.String("project", ".", "project directory")
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("dir=%s\n", *projectDir)

	blobMap := map[string]string{}
	finalBuildsDir := filepath.Join(*projectDir, ".final_builds")
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
			for _, b := range pkg.Builds {
				blobMap[b.Sha1] = b.BlobstoreId
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(blobMap)

	for i, arg := range flag.Args() {
		fmt.Printf("%d: %s\n", i, arg)

		tmp, err := os.CreateTemp("", "release-")
		if err != nil {
			panic(err)
		}
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()
		fmt.Printf("tmp=%s\n", tmp.Name())

		resp, err := http.Get(arg)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		n, err := io.Copy(tmp, resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Bytes copied: %d\n", n)

		_, err = tmp.Seek(0, io.SeekStart)
		if err != nil {
			panic(err)
		}

		gz, err := gzip.NewReader(tmp)
		if err != nil {
			panic(err)
		}

		tarFile := tar.NewReader(gz)
		for {
			h, err := tarFile.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}

			if "release.MF" == h.Name {
				meta := manifest.ReleaseManifest{}
				decoder := yaml.NewDecoder(tarFile)
				err = decoder.Decode(&meta)
				if err != nil {
					panic(err)
				}
				fmt.Println(meta)
			}
		}
	}
}
