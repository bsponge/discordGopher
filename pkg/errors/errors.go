package errors

import "errors"

func BaseError(err error) error {
	unwraped := errors.Unwrap(err)
	if unwraped == nil {
		return err
	}

	return BaseError(unwraped)
}
