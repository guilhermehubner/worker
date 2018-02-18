package errors

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
)

func TestErrorMessage(t *testing.T) {
	customError := errors.New("fail")

	tt := []struct {
		Err         *Error
		Expectation string
	}{
		{
			Err:         &Error{Code: FormatError, Message: "empty value"},
			Expectation: fmt.Sprintf("[Error %d] empty value", FormatError),
		},
		{
			Err:         &Error{Code: ServiceError, Message: "not found"},
			Expectation: fmt.Sprintf("[Error %d] not found", ServiceError),
		},
		{
			Err:         (&Error{Code: FormatError, Message: "nil pointer"}).WithValue(customError),
			Expectation: fmt.Sprintf("[Error %d] nil pointer: fail", FormatError),
		},
		{
			Err:         (&Error{Code: ServiceError, Message: "mysql error"}).WithValue(customError),
			Expectation: fmt.Sprintf("[Error %d] mysql error: fail", ServiceError),
		},
	}

	for _, tc := range tt {
		if tc.Err.Error() != tc.Expectation {
			t.Errorf("expect '%s', but got '%s'", tc.Expectation, tc.Err)
		}
	}
}
