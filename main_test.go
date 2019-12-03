package main

import (
	"bytes"
	"fmt"
	"testing"

	a "github.com/stretchr/testify/assert"

	"github.com/szabado/cache/persistence"
)

func setup(assert *a.Assertions) {
	assert.NoError(persistence.NewFsPersister().Wipe())
}

func TestRunRoot(t *testing.T) {
	testCases := []struct {
		input  []string
		output string
		error  bool
	}{
		{
			input:  []string{"cache", "--verbose", "echo", `-e`, `test\t`},
			output: "test\t\n",
		},
		{
			input:  []string{"cache", "--clean"},
			output: "",
		},
		{
			input: []string{"cache", "--verbose"},
			error: true,
		},
		{
			input:  []string{"cache", "echo"},
			output: "\n",
		},
	}

	for i, test := range testCases {
		t.Run(fmt.Sprint(i, test.input), func(t *testing.T) {
			assert := a.New(t)
			setup(assert)

			var buf bytes.Buffer
			err := runRoot(test.input, &buf)
			if test.error {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(test.output, buf.String())

			buf.Reset()
			_, _, cmd, err := parseArgs(test.input)
			assert.NoError(err)
			persistence.NewFsPersister().ReadInto(cmd, &buf)
			assert.Equal(test.output, buf.String())
		})
	}
}

func TestParseArgs(t *testing.T) {
	testCases := []struct {
		input      []string
		verbose    bool
		clearCache bool
		output     []string
	}{
		{
			input:      []string{"cache", "echo", `-e`, `test\t`},
			verbose:    false,
			clearCache: false,
			output:     []string{"echo", "-e", "test\\t"},
		},
		{
			input:      []string{"cache", "--clean"},
			verbose:    false,
			clearCache: true,
			output:     nil,
		},
		{
			input:      []string{"cache", "--verbose"},
			verbose:    true,
			clearCache: false,
			output:     nil,
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

func TestEscape(t *testing.T) {
	testCases := []struct {
		input  []string
		output string
	}{
		{
			input:  []string{"a", "b"},
			output: "a b ",
		},
	}

	for i, test := range testCases {
		t.Run(fmt.Sprint(i, test.input), func(t *testing.T) {
			assert := a.New(t)
			setup(assert)

			output := escapeAndJoin(test.input)
			assert.Equal(test.output, output)
		})
	}
}
