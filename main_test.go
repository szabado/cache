package main

import (
	"fmt"
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
	a "github.com/stretchr/testify/assert"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func setup(assert *a.Assertions) {
	db, _ := badger.Open(badgerOptions)
	assert.NoError(nukeDatabase(db))
}

func TestRunRoot(t *testing.T) {
	testCases := []struct{
		input []string
		output string
		error bool
	}{
		{
			input: []string{"cache", "echo", `-e`, `test\t`},
			output: "test\t\n",
		},
		{
			input: []string{"cache", "--clean"},
			output: "",
		},
		{
			input: []string{"cache", "--verbose"},
			error: true,
		},
	}

	for i, test := range testCases {
		t.Run(fmt.Sprint(i, test.input), func(t *testing.T) {
			assert := a.New(t)
			setup(assert)

			output, err := runRoot(test.input)
			if test.error {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(test.output, string(output))
		})
	}
}

func TestParseArgs(t *testing.T) {
	testCases := []struct{
		input []string
		verbose bool
		clearCache bool
		output []string
	}{
		{
			input: []string{"cache", "echo", `-e`, `test\t`},
			verbose: false,
			clearCache: false,
			output: []string{"echo", "-e", "test\\t"},
		},
		{
			input: []string{"cache", "--clean"},
			verbose: false,
			clearCache: true,
			output: nil,
		},
		{
			input: []string{"cache", "--verbose"},
			verbose: true,
			clearCache: false,
			output: nil,
		},

	}

	for i, test := range testCases {
		t.Run(fmt.Sprint(i, test.input), func(t *testing.T) {
			assert := a.New(t)
			setup(assert)

			verbose, clearCache, output, err := parseArgs(test.input)
			assert.NoError(err)
			assert.Equal(test.verbose, verbose)
			assert.Equal(test.clearCache, clearCache)
			assert.Equal(test.output, output)
		})
	}
}