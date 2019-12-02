package persistence

import (
	"crypto/sha1"
	"encoding/base64"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
)

const directory = "/tmp/cache-fsdb"

type FsPersister struct {
	// TODO: Add TTL
}

var ErrKeyNotFound = errors.New("Key not found")

func getFilepath(key []byte) string {
	hasher := sha1.New()
	hasher.Write(key)
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return filepath.Join(directory, hash)
}

func (p *FsPersister) ReadInto(key []byte, target io.Writer) error {
	path := getFilepath(key)
	file, err := os.Open(path)
	if err != nil {
		return ErrKeyNotFound
	}
	defer file.Close()

	_, err = io.Copy(target, file)
	return err
}

func (p *FsPersister) GetFileForKey(key []byte) (*os.File, error) {
	os.MkdirAll(directory, 0700)
	return os.OpenFile(getFilepath(key), os.O_RDWR|os.O_CREATE, 0600)
}

func (p *FsPersister) Wipe() error {
	return os.RemoveAll(directory)
}

func NewFsPersister() *FsPersister {
	return &FsPersister{}
}
