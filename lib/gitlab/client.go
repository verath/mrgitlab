package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
)

// Client is a client for the GitLab v4 REST api.
type Client struct {
	logger       *logrus.Entry
	httpClient   *http.Client
	baseURL      *url.URL
	privateToken string
}

// NewClient creates a new Client. The rawBaseURL should point to the GitLab
// server, e.g. "https://gitlab.com/". The privateToken is the GitLab private
// token that the client should use for its requests, see
// https://docs.gitlab.com/ee/api/README.html#private-tokens
func NewClient(logger *logrus.Logger, rawBaseURL string, privateToken string) (*Client, error) {
	logEntry := logger.WithField("module", "gitlab")
	apiURL := rawBaseURL + "/api/v4/"
	baseURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, errors.Wrapf(err, "Error parsing apiURL: %s", apiURL)
	}
	if privateToken == "" {
		logger.Warn("Using an empty GitLab privateToken, this will likely make all GitLab update requests fail!")
	}
	return &Client{
		logger:       logEntry,
		httpClient:   &http.Client{},
		baseURL:      baseURL,
		privateToken: privateToken,
	}, nil
}

// newRequest creates a new http.request with the given method. The path parameter
// is resolved against the client's baseURL. The body, if provided, is JSON-encoded
// and appended to the request. The request will have the required privateToken added
// as a header.
func (c *Client) newRequest(ctx context.Context, method string, path string, body interface{}) (*http.Request, error) {
	// Resolve path against the baseURL
	u, err := url.Parse(path)
	if err != nil {
		return nil, errors.Wrapf(err, "Error parsing path: %s", path)
	}
	reqURL := c.baseURL.ResolveReference(u)
	// If we have a body, encode it as JSON
	buf := &bytes.Buffer{}
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, errors.Wrap(err, "Error encoding body as JSON")
		}
	}
	req, err := http.NewRequest(method, reqURL.String(), buf)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating Request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("PRIVATE-TOKEN", c.privateToken)
	return req.WithContext(ctx), nil
}

// checkResponse returns an error if the response has an error code
// that is outside the "good" 200-300 range.
func (c *Client) checkResponse(res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}
	return errors.Errorf("Bad response code: %d", res.StatusCode)
}

// AddMergeRequestNote creates a new note on the merge request identified
// by the mergeRequestID. It returns an error if the request was not
// successful, or if the context was cancelled.
func (c *Client) AddMergeRequestNote(ctx context.Context, mergeRequestID MergeRequestID, note *Note) error {
	path := fmt.Sprintf("projects/%d/merge_requests/%d/notes", mergeRequestID.ProjectID, mergeRequestID.IID)
	req, err := c.newRequest(ctx, "POST", path, note)
	if err != nil {
		return errors.Wrap(err, "Error creating request")
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "Error sending request")
	}
	defer func() {
		c.logger.Debugf("POST %s - %d", req.URL, res.StatusCode)
	}()
	defer res.Body.Close()
	if err := c.checkResponse(res); err != nil {
		return errors.Wrap(err, "Bad response")
	}
	return nil
}
