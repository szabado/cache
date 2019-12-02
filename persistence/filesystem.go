package persistence

import (
	"crypto/sha1"
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger"
)

const directory = "/tmp/cache-fsdb"

var _ Persister = (*fsPersister)(nil)

type fsPersister struct {
}

func getFilepath(key []byte) string {
	hasher := sha1.New()
	hasher.Write(key)
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return filepath.Join(directory, hash)
}

func (p *fsPersister) ReadInto(key []byte, target io.Writer) error {
	path := getFilepath(key)
	if _, err := os.Stat(path); err != nil {
		// TODO: Fix this hack
		return badger.ErrKeyNotFound
	}

	bytes, err := ioutil.ReadFile(path)

	target.Write(bytes)
	return err
}

func (p *fsPersister) Persist(key, value []byte) error {
	path := getFilepath(key)
	if err := os.MkdirAll(directory, os.FileMode(0700)); err != nil {
		return err
	}

	return ioutil.WriteFile(path, value, os.FileMode(0600))
}

func (p *fsPersister) Wipe() error {
	return os.RemoveAll(directory)
}

func (p *fsPersister) Close() error {
	// noop
	return nil
}

func NewFsPersister() Persister {
	return &fsPersister{}
}
