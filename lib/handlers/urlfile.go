package handlers

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/verath/mrgitlab/lib/gitlab"
)

// NewURLFile creates a new MergeRequestHandlerFunc that send a GET request to the
// given URL and, given the request is successful, returns it as the the
// handler's message.
func NewURLFile(fileURL string) MergeRequestHandlerFunc {
	// Use a custom http client that does not follow redirects. We
	// expect to be given a direct link to the file, not a redirect.
	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return MergeRequestHandlerFunc(func(ctx context.Context, webhook *gitlab.MergeRequestWebhook) (string, error) {
		req, err := http.NewRequest("GET", fileURL, nil)
		if err != nil {
			return "", errors.Wrapf(err, "error creating request for url: %s", fileURL)
		}
		resp, err := httpClient.Do(req.WithContext(ctx))
		if err != nil {
			return "", errors.Wrap(err, "error performing request")
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return "", errors.Errorf("bad status code: %d", resp.StatusCode)
		}
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", errors.Wrap(err, "could not read response body")
		}
		return string(contents), nil
	})
}
