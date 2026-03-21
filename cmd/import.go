package cmd

import (
	"archive/tar"
	"bosh-blob-experiment/blobstore"
	"bosh-blob-experiment/manifest"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	cli "github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

// ImportRelease will read through the TGZ archive and check if the blob exists in the blobstore, copying anything that is not present.
func ImportRelease(_ context.Context, cmd *cli.Command) error {
	projectDir := cmd.String("project")

	for i := 0; i < cmd.Args().Len(); i++ {
		arg := cmd.Args().Get(i)
		fmt.Printf("%d: %s\n", i, arg)

		tmp, err := os.CreateTemp("", "release-")
		if err != nil {
			return err
		}
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		var reader io.Reader
		u, err := url.Parse(arg)
		if err == nil && u.Scheme != "" {
			log.Printf("downloading '%s'...", arg)
			resp, err := http.Get(arg)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			reader = resp.Body
		} else {
			log.Printf("reading '%s'...", arg)
			f, err := os.OpenFile(arg, os.O_RDONLY, os.ModePerm)
			if err != nil {
				return err
			}
			defer f.Close()
			reader = f
		}

		n, err := io.Copy(tmp, reader)
		if err != nil {
			return err
		}
		log.Printf("bytes copied: %d\n", n)

		_, err = tmp.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		gz, err := gzip.NewReader(tmp)
		if err != nil {
			return err
		}

		model, err := manifest.Load(projectDir)
		if err != nil {
			return err
		}

		blobstore, _, err := blobstore.NewFromConfig(projectDir)
		if err != nil {
			return err
		}

		tarFile := tar.NewReader(gz)
		meta := manifest.ReleaseManifest{}
		for {
			h, err := tarFile.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			// Note: We assume "release.MF" is _first_ in the archive
			if "release.MF" == h.Name {
				decoder := yaml.NewDecoder(tarFile)
				err = decoder.Decode(&meta)
				if err != nil {
					return err
				}
				log.Printf("'release.MF' read: %v", meta)
			} else if h.Typeflag == tar.TypeReg {
				found := false
				for _, job := range meta.Jobs {
					if fmt.Sprintf("jobs/%s.tgz", job.Name) == h.Name {
						found = true
						err = handleBlob(model, blobstore, tarFile, "job", job.Name, job.Version)
						if err != nil {
							return err
						}
					}
				}
				for _, pkg := range meta.Packages {
					if fmt.Sprintf("packages/%s.tgz", pkg.Name) == h.Name {
						found = true
						err = handleBlob(model, blobstore, tarFile, "package", pkg.Name, pkg.Version)
						if err != nil {
							return err
						}
					}
				}
				if !found {
					log.Printf("file '%s' is in the tar file but not in the 'release.MF' descriptor", h.Name)
				}
			}
		}
	}
	return nil
}

func handleBlob(model *manifest.Model, blobstore blobstore.Blobstore, reader io.Reader, label, name, version string) error {
	build := model.FindBuildByVersion(version)
	if build == nil {
		log.Printf("warning: unable to locate blob for %s named '%s'", label, name)
		return nil
	}
	exists, err := blobstore.Exists(build.BlobstoreId)
	if err != nil {
		return nil
	}
	if exists {
		log.Printf("blob version %s (%s %s) exists", version, label, name)
	}
	if !exists {
		log.Printf("blob version %s (%s %s) does not exist; extracting", version, label, name)
		tmp, err := os.CreateTemp("", "blob-")
		if err != nil {
			return err
		}
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		n, err := io.Copy(tmp, reader)
		if err != nil {
			return err
		}
		log.Printf("copied %d bytes from '%s' to '%s'", n, name, tmp.Name())

		_, err = tmp.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		err = blobstore.Put(tmp, build.BlobstoreId)
		if err != nil {
			return err
		}
	}
	return nil
}
