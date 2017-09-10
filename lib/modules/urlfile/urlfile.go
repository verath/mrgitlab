package urlfile

import (
	"context"
	"github.com/pkg/errors"
	"github.com/verath/mrgitlab/lib"
	"github.com/verath/mrgitlab/lib/gitlab"
	"io/ioutil"
	"net/http"
)

// New creates a new MergeRequestHandler that send a GET request to the
// given URL and, given the request is successful, returns it as the the
// handler's message.
func New(fileURL string) mrgitlab.MergeRequestHandler {
	return mrgitlab.MergeRequestHandlerFunc(func(ctx context.Context, webhook *gitlab.MergeRequestWebhook) (string, error) {
		req, err := http.NewRequest("GET", fileURL, nil)
		if err != nil {
			return "", errors.Wrapf(err, "error creating request for url: %s", fileURL)
		}
		resp, err := http.DefaultClient.Do(req.WithContext(ctx))
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
