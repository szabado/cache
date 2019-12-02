package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/szabado/cache/persistence"
)

var usage = `cache: A Cache for slow shell commands.

Querying log clusters or curling API endpoints can have a latency that can
make it annoying to build up a pipe pipeline iteratively. This tool caches
those results for you so you iterate quickly.

cache runs the command for you and stores the result, and then returns the
output to you. Any data stored has a TTL of 1 hour, and subsequent calls of
the same command will return the stored result. cache will only store the
results of successful commands: if your bash command has a non-zero exit
code, then it will be uncached.

Usage:
  cache [flags] [command]

Flags:
      --clear, --clean   Clear the cache.
  -v, --verbose          Verbose logging.

Examples

  cache curl -X GET example.com
`

func main() {
	err := runRoot(os.Args, os.Stdout)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", err)
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
}

func parseArgs(args []string) (verbose bool, clearCache bool, command []string, err error) {
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--clean", "--clear":
			clearCache = true
		case "--verbose", "-v":
			verbose = true
		default:
			if strings.HasPrefix(arg, "-") {
				return false, false, nil, errors.Errorf("unknown flag: %s", arg)
			}
			return verbose, clearCache, args[i:], nil
		}
	}

	return verbose, clearCache, nil, nil
}

func runRoot(args []string, output io.Writer) error {
	verbose, clearCache, command, err := parseArgs(args)
	if err != nil {
		return err
	}
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.ErrorLevel)
	}

	if len(command) == 0 && !clearCache {
		return errors.New("No arguments provided")
	}

	logrus.Infof("Command: %s (%#[1]v)", command)
	persister := persistence.NewFsPersister()
	if persister == nil {
		logrus.WithError(err).Errorf("failed to open database, not caching execution")
	} else {
		defer persister.Close()
	}

	if clearCache {
		logrus.Info("Deleting database")
		return persister.Wipe()
	}

	return runCommand(persister, command, output)
}

func runCommand(persister persistence.Persister, command []string, output io.Writer) error {
	var (
		commandKey []byte = []byte(strings.Join(command, " "))
		cmdOutput  []byte
		err        error
	)

	if persister != nil {
		err = persister.ReadInto(commandKey, output)

		if err != nil && err != badger.ErrKeyNotFound {
			logrus.WithError(err).Errorf("Unknown error trying to find previous execution")
		} else if err == nil {
			logrus.Debug("Found previous execution, exiting early")
			return nil
		}
	}

	logrus.Debugf("Failed to find previous execution, executing command")
	cmdOutput, err = executeCommand(command)
	output.Write(cmdOutput)
	if err != nil {
		return err
	}

	if persister != nil {
		logrus.Debugf("Persisting command output")
		err := persister.Persist(commandKey, cmdOutput)
		if err != nil {
			logrus.WithError(err).Warn("Failed to persist output")
		}
	}
	return nil
}

func executeCommand(command []string) ([]byte, error) {
	logrus.Infof("Executing command: %s (%#[1]v)", command)
	cmd := exec.Command("bash", "-c", escape(command))
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

func escape(cmdSegments []string) string {
	var builder strings.Builder
	for _, seg := range cmdSegments {
		builder.WriteString(shellescape.Quote(seg))
		builder.WriteByte(' ')
	}

	return builder.String()
}
