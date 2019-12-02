package persistence

import "io"

type Persister interface {
	ReadInto(key []byte, target io.Writer) error
	Persist(key, value []byte) error
	Wipe() error
	Close() error
}
