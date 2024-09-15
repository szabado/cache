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
		err    bool
	}{
		{
			input:  []string{"cache", "--verbose", "echo", "-e", `test\t`},
			output: "test\t\n",
		},
		{
			input:  []string{"cache", "--clean"},
			output: "",
		},
		{
			input: []string{"cache", "--verbose"},
			err:   true,
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
			if test.err {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(test.output, buf.String())

			// Don't bother writing to the file if an error is expected
			if !test.err {
				buf.Reset()
				// Test the result matches after writing to and reading from the file
				_, _, _, cmd, err := parseArgs(test.input)
				assert.NoError(err)
				persistence.NewFsPersister().ReadInto(cmd, &buf)
				assert.Equal(test.output, buf.String())
			}
		})
	}
}

func TestCommandIsCached(t *testing.T) {
	assert := a.New(t)

	setup(assert)

	buffer := bytes.NewBufferString("")
	assert.NoError(runRoot([]string{"cache", "mktemp"}, buffer))
	firstResult := buffer.String()

	buffer.Reset()
	assert.NoError(runRoot([]string{"cache", "mktemp"}, buffer))
	secondResult := buffer.String()

	fmt.Println(firstResult)

	assert.Equal(firstResult, secondResult)
}

func TestParseArgs(t *testing.T) {
	testCases := []struct {
		input      []string
		verbose    bool
		clearCache bool
		overwrite  bool
		err        bool
		command    string
	}{
		{
			input:      []string{"cache", "echo", "-e", `test\t`},
			verbose:    false,
			clearCache: false,
			overwrite:  false,
			err:        false,
			command:    "echo -e 'test\\t' ",
		},
		{
			input:      []string{"cache", "--clean"},
			verbose:    false,
			clearCache: true,
			overwrite:  false,
			err:        false,
			command:    "",
		},
		{
			input:      []string{"cache", "--verbose", "echo", "-e", `test\t`},
			verbose:    true,
			clearCache: false,
			overwrite:  false,
			err:        false,
			command:    "echo -e 'test\\t' ",
		},
		{
			input:      []string{"cache", "--overwrite", "echo", "-e", `test\t`},
			verbose:    false,
			clearCache: false,
			overwrite:  true,
			err:        false,
			command:    "echo -e 'test\\t' ",
		},
		{
			input:      []string{"cache", "--verbose"},
			verbose:    true,
			clearCache: false,
			overwrite:  false,
			err:        true,
			command:    "",
		},
	}

	for i, test := range testCases {
		t.Run(fmt.Sprint(i, test.input), func(t *testing.T) {
			assert := a.New(t)
			setup(assert)

			verbose, clearCache, overwrite, command, err := parseArgs(test.input)
			if test.err {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(test.verbose, verbose)
			assert.Equal(test.clearCache, clearCache)
			assert.Equal(test.command, command)
			assert.Equal(test.overwrite, overwrite)
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
