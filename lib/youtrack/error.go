package youtrack

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// apiHTTPError is a custom error we return for http request
// that are not of an expected error code.
type apiHTTPError struct {
	statusCode int
}

// newAPIHTTPError returns a new apiHTTPError for the given response.
func newAPIHTTPError(res *http.Response) apiHTTPError {
	return apiHTTPError{statusCode: res.StatusCode}
}

// Error implements the error interface
func (err apiHTTPError) Error() string {
	return fmt.Sprintf("bad status code: %d", err.statusCode)
}

// IsHTTPStatusError returns true if the cause of the given error
// was that the YouTrack API responded with the given statusCode.
func IsHTTPStatusError(err error, statusCode int) bool {
	apiErr, ok := errors.Cause(err).(apiHTTPError)
	return ok && apiErr.statusCode == statusCode
}
