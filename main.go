package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
var (
	badgerOptions = badger.DefaultOptions("/tmp/cache-database")
)

func init() {
	badgerOptions.Logger = logrus.StandardLogger()
}

func main() {
	output, err := runRoot(os.Args)
	fmt.Printf("%s", output)
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

func runRoot(args []string) ([]byte, error) {
	verbose, clearCache, command, err := parseArgs(args)
	if err != nil {
		return nil, err
	}
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.ErrorLevel)
	}

	if len(command) <= 1 && !clearCache {
		return nil, errors.New("No arguments provided")
	}

	logrus.Infof("Command: %s", command)
	db, err := badger.Open(badgerOptions)
	if err != nil {
		logrus.WithError(err).Errorf("failed to open database, not caching execution")
		db = nil
	} else {
		defer db.Close()
	}

	if clearCache {
		logrus.Info("Deleting database")
		return nil, nukeDatabase(db)
	}

	return runCommand(db, command)
}

func nukeDatabase(db *badger.DB) error {
	if db != nil {
		return db.DropAll()
	} else {
		logrus.Info("No database connection, trying to delete directory.")
		return os.RemoveAll(badgerOptions.Dir)
	}
}

func runCommand(db *badger.DB, command []string) ([]byte, error) {
	var (
		output []byte
		err error
	)

	if db != nil {
		output, err = fetch(db, command)

		if err != nil && err != badger.ErrKeyNotFound {
			logrus.WithError(err).Errorf("Unknown error trying to find previous execution")
		} else if err == nil {
			logrus.Debug("Found previous execution, exiting early")
			return output, nil
		}
	}

	logrus.Debugf("Failed to find previous execution, executing command")
	output, err = executeCommand(command)
	if err != nil {
		return nil, err
	}

	if db != nil {
		persist(db, command, output)
	}
	return output, nil
}

func fetch(db *badger.DB, key []string) ([]byte, error) {
	var (
		joinedKey = []byte(strings.Join(key, " "))
		output []byte
	)

	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(joinedKey)
		if err != nil {
			return err
		}

		output, err = item.ValueCopy(nil)
		return err
	})

	return output, err
}

func executeCommand(command []string) ([]byte, error) {
	logrus.Infof("Executing command: %s", command)
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

func persist(db *badger.DB, key []string, value []byte) {
	var joinedKey = []byte(strings.Join(key, " "))
	err := db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(joinedKey, value).WithTTL(time.Hour)
		return txn.SetEntry(entry)
	})

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"key": key,
			"value": value,
		}).Warn("Failed to persist data")
	}
}
