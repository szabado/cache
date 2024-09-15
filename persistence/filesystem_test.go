package persistence

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	a "github.com/stretchr/testify/assert"
)

func setup(assert *a.Assertions) {
	assert.NoError(NewFsPersister().Wipe())
}

func TestWithinTTL(t *testing.T) {
	assert := a.New(t)
	setup(assert)

	filename := "./ImATestFile.out"
	file, err := os.Create(filename)
	assert.NoError(err)
	defer os.Remove(filename)
	assert.NoError(file.Close())

	assert.NoError(os.Chtimes(filename, time.Now(), time.Now()))
	file, err = os.Open(filename)
	assert.NoError(err)
	assert.True(isWithinTTL(file))
	assert.NoError(file.Close())
}

func TestOutsideTTL(t *testing.T) {
	assert := a.New(t)
	setup(assert)

	filename := "./ImATestFile.out"
	file, err := os.Create(filename)
	assert.NoError(err)
	defer os.Remove(filename)
	assert.NoError(file.Close())

	assert.NoError(os.Chtimes(filename, time.Now(), time.Now().Add(-1*time.Hour-1*time.Second)))
	file, err = os.Open(filename)
	assert.NoError(err)
	assert.False(isWithinTTL(file))
	assert.NoError(file.Close())
}

func TestWipe(t *testing.T) {
	assert := a.New(t)
	setup(assert)

	testfile := "testfile"
	fs := NewFsPersister()

	writer, err := fs.GetWriterForKey(testfile)
	assert.NoError(err)

	_, err = writer.Write([]byte("test"))
	assert.NoError(err)

	assert.NoError(writer.Close())

	buffer := bytes.NewBufferString("")
	fs.ReadInto(testfile, buffer)

	assert.Equal("test", buffer.String())

	buffer.Reset()

	assert.NoError(fs.Wipe())
	err = fs.ReadInto(testfile, buffer)
	assert.Equal(ErrKeyNotFound, err)
	assert.Equal("", buffer.String())
}

func TestExpiredFile(t *testing.T) {
	assert := a.New(t)
	setup(assert)
	testfile := "testfile"

	fs := NewFsPersister()

	writer, err := fs.GetWriterForKey(testfile)
	assert.NoError(err)

	writer.Write([]byte("test"))

	entries, err := os.ReadDir(directory)
	assert.NoError(err)
	assert.Equal(1, len(entries))
	hashedFileName := filepath.Join(directory, entries[0].Name())

	assert.NoError(os.Chtimes(hashedFileName, time.Now(), time.Now().Add(-1*time.Hour-1*time.Second)))

	buffer := bytes.NewBufferString("")
	err = fs.ReadInto(testfile, buffer)

	assert.Equal(ErrKeyNotFound, err)
	assert.Equal("", buffer.String())
}
