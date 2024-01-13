package errs

import "errors"

var (
	ErrRepoCreate          error = errors.New("failed to create repository instance")
	ErrUnsupportedCurrency error = errors.New("unsupported currency")
)
