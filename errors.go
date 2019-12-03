package main

import "github.com/pkg/errors"

var _ error = (*UsageError)(nil)

type UsageError struct {
	error
}

func NewUsageError(err error) error {
	return &UsageError{
		errors.Wrap(err, "usage error"),
	}
}
