package persistence

import (
	"io"
	"os"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
)

var _ Persister = (*badgerPersister)(nil)

var BadgerOptions = badger.DefaultOptions("/tmp/cache-badgerdb")

func init() {
	BadgerOptions.Logger = logrus.StandardLogger()
}

type badgerPersister struct {
	db *badger.DB
}

func (p *badgerPersister) ReadInto(key []byte, target io.Writer) error {
	return p.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		item.Value(func(val []byte) error {
			_, err := target.Write(val)
			return err
		})
		return err
	})
}

func (p *badgerPersister) Persist(key, value []byte) error {
	return p.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, value).WithTTL(time.Hour)
		return txn.SetEntry(entry)
	})
}

func (p *badgerPersister) Wipe() error {
	if err := p.db.DropAll(); err != nil {
		logrus.Warn("No database connection, trying to delete directory.")
		err = os.RemoveAll(BadgerOptions.Dir)
		return err
	}
	return nil
}

func (p *badgerPersister) Close() error {
	return p.db.Close()
}

func NewBadgerDbPersister() Persister {
	db, err := badger.Open(BadgerOptions)
	if err != nil {
		return nil
	}
	return &badgerPersister{
		db: db,
	}
}
