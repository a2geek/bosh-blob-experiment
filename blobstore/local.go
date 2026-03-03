package blobstore

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

func NewLocalBlobstore(blobstoreDir string) Blobstore {
	return &localBlobstore{
		blobstoreDir: blobstoreDir,
	}
}

type localBlobstore struct {
	blobstoreDir string
}

func (l *localBlobstore) Exists(name string) (bool, error) {
	_, err := os.Stat(filepath.Join(l.blobstoreDir, name))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (l *localBlobstore) Get(src string, dest io.WriterAt) error {
	panic("unimplemented")
}

func (l *localBlobstore) List() ([]string, error) {
	panic("unimplemented")
}

func (l *localBlobstore) Put(src io.ReadSeeker, dest string) error {
	file, err := os.OpenFile(filepath.Join(l.blobstoreDir, dest), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, src)
	return err
}
