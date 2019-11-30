package cmd

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	verbose       bool
	clearCache    bool
	badgerOptions = badger.DefaultOptions("/tmp/cache-database")
)

func init() {
	RootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose logging.")
	RootCmd.Flags().BoolVar(&clearCache, "clear", false, "Clear the cache")
}

var RootCmd = &cobra.Command{
	Use:                "cache [command]",
	FParseErrWhitelist: struct{ UnknownFlags bool }{UnknownFlags: true},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 && !clearCache {
			return errors.New("Expected at least one argument, got 0")
		}
		return nil
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		badgerOptions.Logger = logrus.StandardLogger()

		if !verbose {
			logrus.SetLevel(logrus.FatalLevel)
		} else {
			logrus.SetLevel(logrus.DebugLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		output, err := runRoot(args)
		_, _ = os.Stdout.Write(output)

		if err != nil {
			os.Exit(1)
		}
	},
}

func runRoot(args []string) ([]byte, error) {
	var command = strings.Join(args, " ")

	db, err := badger.Open(badgerOptions)
	if err != nil {
		logrus.WithError(err).Errorf("failed to open database, not caching execution")
		db = nil
	}

	if clearCache {
		if db == nil {
			return nil, errors.New("could not open database to clear it")
		}
		return nil, db.DropAll()
	}

	return runCommand(db, command)
}

func runCommand(db *badger.DB, command string) ([]byte, error) {
	var (
		output []byte
		err error
	)

	if db != nil {
		output, err = fetch(db, command)

		if err != nil && err != badger.ErrKeyNotFound {
			logrus.WithError(err).Errorf("Unknown error trying to find previous execution, not caching execution")
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

func fetch(db *badger.DB, key string) ([]byte, error) {
	var output []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		output, err = item.ValueCopy(nil)
		return err
	})

	return output, err
}

func executeCommand(command string) ([]byte, error) {
	logrus.Info("Executing command %s", command)
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

func persist(db *badger.DB, key string, value []byte) {
	err := db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), value).WithTTL(time.Hour)
		return txn.SetEntry(entry)
	})

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"key": key,
			"value": value,
		}).Warn("Failed to persist data")
	}
}