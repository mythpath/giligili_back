package cmdb

import (
	"errors"
	"strings"
)

var (
	// it record not found error, happens when haven't find any matched data when looking up with a struct
	ErrRecordNotFound = errors.New("record not found")
	// invalid SQL error, happens when you passed invalid SQL
	ErrInvalidSQL = errors.New("invalid SQL")
	// unaddressable value
	ErrUnaddressable = errors.New("using unaddressable value")

	// invalid length
	ErrResultSize = errors.New("invalid result size")
	// invalid element type
	ErrElementType = errors.New("invalid element type")
	// server handle error
	ErrHandleFailed = errors.New("handle failed")
)

// Errors contains all happened errors
type Errors []error

// IsRecordNotFoundError returns current error has record not found error or not
func IsRecordNotFoundError(err error) bool {
	if errs, ok := err.(Errors); ok {
		for _, err := range errs {
			if errors.Is(err, ErrRecordNotFound) {
				return true
			}
		}
	}
	return errors.Is(err, ErrRecordNotFound)
}

// GetErrors gets all happened errors
func (errs Errors) GetErrors() []error {
	return errs
}

// Add adds an error
func (errs Errors) Add(newErrors ...error) Errors {
	for _, err := range newErrors {
		if err == nil {
			continue
		}

		if errors, ok := err.(Errors); ok {
			errs = errs.Add(errors...)
		} else {
			ok = true
			for _, e := range errs {
				if err == e {
					ok = false
				}
			}
			if ok {
				errs = append(errs, err)
			}
		}
	}
	return errs
}

// Error format happened errors
func (errs Errors) Error() string {
	var errors = []string{}
	for _, e := range errs {
		errors = append(errors, e.Error())
	}
	return strings.Join(errors, "; ")
}
