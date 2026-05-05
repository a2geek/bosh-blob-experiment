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
	"strings"

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
		defer deleteTempFile(tmp)

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

		// This is the (possibly) remote blobstore -or- the final_blobs directory
		remoteBlobstore, _, err := blobstore.NewFromConfig(projectDir)
		if err != nil {
			return err
		}

		// This is for any blobs added and managed via "add-blobs" that don't get synchronized.
		localBlobstoreConfig := &blobstore.FinalManifest{
			Blobstore: blobstore.FinalBlobstore{
				Provider: "local",
				Options: map[string]any{
					"blobstore_path": "blobs",
				},
			},
		}
		localBlobstore, err := blobstore.NewFromManifest(projectDir, localBlobstoreConfig)
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
						tmp, err = handleBlob(model, remoteBlobstore, tarFile, "job", job.Name, job.Version)
						if err != nil {
							return err
						}
						defer deleteTempFile(tmp)
					}
				}
				for _, pkg := range meta.Packages {
					if fmt.Sprintf("packages/%s.tgz", pkg.Name) == h.Name {
						found = true
						tmp, err = handleBlob(model, remoteBlobstore, tarFile, "package", pkg.Name, pkg.Version)
						if err != nil {
							return err
						}
						// check contents and possibly copy into localBlobstore
						if tmp != nil {
							defer deleteTempFile(tmp)
							err = lookForLocalBlobs(model, localBlobstore, tmp)
							if err != nil {
								return err
							}
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

func deleteTempFile(file *os.File) {
	if file != nil {
		file.Close()
		os.Remove(file.Name())
	}
}

func handleBlob(model *manifest.Model, blobstore blobstore.Blobstore, reader io.Reader, label, name, version string) (*os.File, error) {
	build := model.FindBuildByVersion(version)
	if build == nil {
		log.Printf("warning: unable to locate blob for %s named '%s'", label, name)
		return nil, nil
	}
	exists, err := blobstore.Exists(build.BlobstoreId)
	if err != nil {
		return nil, nil
	}

	// We (lazily) extract the blob -- if it's a package we need to dig around to see if other blobs exist (so it needs to be extracted regardless)
	tmp, err := os.CreateTemp("", "blob-")
	if err != nil {
		return nil, err
	}

	n, err := io.Copy(tmp, reader)
	if err != nil {
		return tmp, err
	}
	log.Printf("copied %d bytes from '%s' to '%s'", n, name, tmp.Name())

	_, err = tmp.Seek(0, io.SeekStart)
	if err != nil {
		return tmp, err
	}

	if exists {
		log.Printf("blob version %s (%s %s) exists", version, label, name)
	} else {
		log.Printf("blob version %s (%s %s) does not exist; extracting", version, label, name)

		err = blobstore.Put(tmp, build.BlobstoreId)
		return tmp, err
	}
	return tmp, nil
}

func lookForLocalBlobs(model *manifest.Model, localBlobstore blobstore.Blobstore, reader io.Reader) error {
	gz, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}

	tarFile := tar.NewReader(gz)
	for {
		h, err := tarFile.Next()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		name, _ := strings.CutPrefix(h.Name, "./")
		if blob, found := model.FindBlobByName(name); found {
			// If an objectId has been assigned, it's in the other blobstore, not the local blobstore.
			if blob.ObjectId != "" {
				exists, err := localBlobstore.Exists(name)
				if err != nil {
					return err
				}
				if !exists {
					log.Printf("blob %s does not exist; extracting", name)
					tmp, err := os.CreateTemp("", "blob-")
					if err != nil {
						return nil
					}
					defer deleteTempFile(tmp)

					n, err := io.Copy(tmp, reader)
					if err != nil {
						return err
					}
					log.Printf("copied %d bytes from '%s' to '%s'", n, name, tmp.Name())

					_, err = tmp.Seek(0, io.SeekStart)
					if err != nil {
						return err
					}

					log.Printf("copying '%s' into local blobstore as '%s'", tmp.Name(), name)
					err = localBlobstore.Put(tmp, name)
					return err
				} else {
					log.Printf("blob %s already exists; skipping", name)
				}
			} else {
				log.Printf("blob '%s' already synchronized as '%s'", name, blob.ObjectId)
			}
		}
	}
}
