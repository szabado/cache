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
		err := runRoot(args)
		if err != nil {
			os.Exit(1)
		}
	},
}

func runRoot(args []string) error {
	var rawCommand = strings.Join(args, " ")
	var rawCommandBytes = []byte(rawCommand)

	db, err := badger.Open(badgerOptions)
	if err != nil {
		logrus.WithError(err).Errorf("failed to open database, not caching execution")
		db = nil
	}

	if clearCache {
		if db == nil {
			return errors.New("could not open database to clear it")
		}
		return db.DropAll()
	}

	if db != nil {
		err = printPreviousExecution(db, rawCommandBytes)

		if err != nil && err != badger.ErrKeyNotFound {
			logrus.WithError(err).Errorf("Unknown error trying to find previous execution, not caching execution")
		} else if err == nil {
			logrus.Debug("Found previous execution, exiting early")
			return nil
		}
	}

	logrus.Debugf("Failed to find previous execution, executing command: %s", rawCommand)
	cmd := exec.Command("bash", "-c", rawCommand)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	os.Stdout.Write(output)

	if db != nil {
		err = db.Update(func(txn *badger.Txn) error {
			entry := badger.NewEntry(rawCommandBytes, output).WithTTL(time.Hour)
			return txn.SetEntry(entry)
		})

		if err != nil {
			logrus.WithError(err).Errorf("Failed to store the command result")
		}
	}
	return nil
}

func printPreviousExecution(db *badger.DB, command []byte) error {
	return db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(command)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			_, err := os.Stdout.Write(val)
			return err
		})
	})

}
