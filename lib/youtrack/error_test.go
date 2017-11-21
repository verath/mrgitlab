package youtrack

import (
	"testing"

	"github.com/pkg/errors"
)

func TestIsHTTPStatusError(t *testing.T) {
	tests := []struct {
		// err is the error we are passing to IsHTTPStatusError
		err error
		// statusCode is the status code we are passing to IsHTTPStatusError
		statusCode int
		// expected is the expected result of IsHTTPStatusError given the
		// paramters we have given it
		expected bool
	}{
		// An error that is not an apiHTTPError does not have
		// a status code, so should never match.
		{
			err:        errors.New("generic error"),
			statusCode: 400,
			expected:   false,
		},
		// apiHTTPError with the same status code as tested against
		{
			err:        apiHTTPError{statusCode: 400},
			statusCode: 400,
			expected:   true,
		},
		// apiHTTPError with another status code than tested against
		{
			err:        apiHTTPError{statusCode: 500},
			statusCode: 400,
			expected:   false,
		},
		// Make sure that we unwrap pkg/errors errors
		{
			err:        errors.Wrap(apiHTTPError{statusCode: 500}, "wrapped apiHTTPError"),
			statusCode: 400,
			expected:   false,
		},
	}

	for _, test := range tests {
		actual := IsHTTPStatusError(test.err, test.statusCode)
		if actual != test.expected {
			t.Errorf(""+
				"IsHTTPStatusError\n"+
				"\t err: %v\n"+
				"\t statusCode: %d\n"+
				"\t expected: %v\n"+
				"\t actual: %v\n",
				test.err, test.statusCode, test.expected, actual)
		}
	}
}
