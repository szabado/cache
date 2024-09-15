package persistence

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var directory = filepath.Join(os.TempDir(), "cache-fsdb")

type FsPersister struct {
	// TODO: Add Custom TTL
}

var ErrKeyNotFound = errors.New("Key not found")

func getFilepath(key []byte) string {
	hasher := sha256.New()
	hasher.Write(key)
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return filepath.Join(directory, hash)
}

func isWithinTTL(file *os.File) bool {
	fd, err := file.Stat()
	if err != nil {
		logrus.WithError(err).Warn("Failed to stat file to check TTL")
		return false
	}

	return fd.ModTime().After(time.Now().Add(-1 * time.Hour))
}

func (p *FsPersister) ReadInto(key string, target io.Writer) error {
	path := getFilepath([]byte(key))
	file, err := os.Open(path)
	if err != nil {
		return ErrKeyNotFound
	}
	defer file.Close()
	if !isWithinTTL(file) {
		_ = os.Remove(path)
		return ErrKeyNotFound
	}

	_, err = io.Copy(target, file)
	return err
}

func (p *FsPersister) GetWriterForKey(key string) (io.WriteCloser, error) {
	os.MkdirAll(directory, 0700)
	return os.OpenFile(getFilepath([]byte(key)), os.O_RDWR|os.O_CREATE, 0600)
}

func (p *FsPersister) Wipe() error {
	return os.RemoveAll(directory)
}

func NewFsPersister() *FsPersister {
	return &FsPersister{}
}
