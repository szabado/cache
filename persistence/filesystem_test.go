package persistence

import (
	"os"
	"testing"
	"time"

	a "github.com/stretchr/testify/assert"
)

func TestWithinTTL(t *testing.T) {
	assert := a.New(t)

	filename := "./ImATestFile.out"
	file, err := os.Create(filename)
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

	filename := "./ImATestFile.out"
	file, err := os.Create(filename)
	defer os.Remove(filename)
	assert.NoError(file.Close())

	assert.NoError(os.Chtimes(filename, time.Now(), time.Now().Add(-1 * time.Hour - 1 * time.Second)))
	file, err = os.Open(filename)
	assert.NoError(err)
	assert.False(isWithinTTL(file))
	assert.NoError(file.Close())
}
