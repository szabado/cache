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
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

func parseArgs(args []string) (verbose bool, clearCache bool, command []string) {
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--clean":
			clearCache = true
		case "--verbose":
			verbose = true
		default:
			return verbose, clearCache, args[i:]
		}
	}

	return verbose, clearCache, nil
}

func runRoot(args []string) ([]byte, error) {
	fmt.Printf("%#v\n", args)
	if len(args) <= 1 {
		return nil, errors.New("No arguments provided")
	}

	verbose, clearCache, command := parseArgs(args)
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.ErrorLevel)
	}

	logrus.Infof("Command: %s", command)
	fmt.Printf("%#v %#v %#v\n", command, verbose, clearCache)
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
	if db == nil {
		logrus.Info("No database connection, trying to delete directory.")
		return os.RemoveAll(badgerOptions.Dir)
	} else {
		return db.DropAll()
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
	var joinedKey = []byte(strings.Join(key, " ")) // TODO: test this
	var output []byte
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
	var joinedKey = []byte(strings.Join(key, " ")) // TODO: test this
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
