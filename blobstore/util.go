package blobstore

import (
	"log"
	"os"
	"slices"
)

// Copy blobs from oldBlobstore into newBlobstore; use force to recopy existing blobs
func Copy(oldBlobstore, newBlobstore Blobstore, force bool) error {
	oldBlobs, err := oldBlobstore.List()
	if err != nil {
		return err
	}
	var newBlobs []string
	if !force {
		newBlobs, err = newBlobstore.List()
		if err != nil {
			return err
		}
	}
	for _, blob := range oldBlobs {
		if slices.Contains(newBlobs, blob) {
			log.Printf("skipping blob '%s' since it already exists", blob)
			continue
		}

		log.Printf("copying blob '%s'", blob)

		tmp, err := os.CreateTemp("", "blob-")
		if err != nil {
			return err
		}
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		err = oldBlobstore.Get(blob, tmp)
		if err != nil {
			return err
		}

		err = newBlobstore.Put(tmp, blob)
		if err != nil {
			return err
		}
	}
	return nil
}
