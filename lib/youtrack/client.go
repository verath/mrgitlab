package youtrack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

// Client is a rest client for YouTrack
type Client struct {
	logger     *logrus.Entry
	baseURL    *url.URL
	httpClient *http.Client

	// loginTicketCh is a channel that acts as a mutex for performing
	// a combined login+request. We use a channel instead of a mutex
	// so that we can select on context cancellation while awaiting
	// a ticket.
	loginTicketCh chan struct{}

	username string
	password string
}

// NewClient creats a new YouTrack API client. The rawBaseURL should point to
// the root of the YouTrack instance, e.g. "http://track.example.com:8080/".
// The username and the password is used to authenticate with the YouTrack API.
func NewClient(logger *logrus.Logger, rawBaseURL string, username string, password string) (*Client, error) {
	logEntry := logger.WithField("module", "youtrack")
	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing apiURL: %s", baseURL)
	}
	if username == "" || password == "" {
		return nil, errors.New("both username and password is required")
	}
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not create cookie jar")
	}
	// Setup the httpClient to use a cookieJar. A cookieJar is required
	// for the client to be able to store the cookie set when loggin in
	httpClient := &http.Client{Jar: cookieJar}
	// Create and add an initial value to the loginTicketCh
	loginTicketCh := make(chan struct{}, 1)
	loginTicketCh <- struct{}{}
	return &Client{
		logger:        logEntry,
		baseURL:       baseURL,
		httpClient:    httpClient,
		loginTicketCh: loginTicketCh,
		username:      username,
		password:      password,
	}, nil
}

// resolvePath resolves a given path against the Client's baseURL.
func (c *Client) resolvePath(path string) (*url.URL, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing path: %s", path)
	}
	return c.baseURL.ResolveReference(u), nil
}

// newRequest creates a new http request, with the provided context and method.
// The path is resolved against the Client's baseURL.
func (c *Client) newRequest(ctx context.Context, method string, path string) (*http.Request, error) {
	reqURL, err := c.resolvePath(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not resolve path: %s", path)
	}
	req, err := http.NewRequest(method, reqURL.String(), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create request for path: %s", path)
	}
	req.Header.Set("Accept", "application/json")
	return req.WithContext(ctx), nil
}

// checkResponse returns an error if the response is not in the
// 200 range.
func (c *Client) checkResponse(res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}
	return newAPIHTTPError(res)
}

// do performs a request with the Client's httpClient
func (c *Client) do(req *http.Request) (*http.Response, error) {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error performing request")
	}
	c.logger.Debugf("%s %s - %d", req.Method, req.URL, res.StatusCode)
	return res, err
}

// doWithLogin performs a login request followed by the provided request, given
// that the login request was successful. doWithLogin additionally synchronizes
// with the loginTicketCh, so that only a single login+req combination is performed
// at a time.
func (c *Client) doWithLogin(ctx context.Context, req *http.Request) (*http.Response, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.loginTicketCh:
	}
	defer func() { c.loginTicketCh <- struct{}{} }()
	if err := c.login(ctx); err != nil {
		return nil, errors.Wrap(err, "could not login")
	}
	return c.do(req)
}

// login performs a login request with the client's username and password.
// A successful login will result in a session cookie being added to the
// httpClient's cookie jar, which will then act as authentication for other
// requests.
func (c *Client) login(ctx context.Context) error {
	path := fmt.Sprintf("rest/user/login?login=%s&password=%s", c.username, c.password)
	req, err := c.newRequest(ctx, "POST", path)
	if err != nil {
		return errors.Wrap(err, "error creating request")
	}
	res, err := c.do(req)
	if err != nil {
		return errors.Wrap(err, "error performing request")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return newAPIHTTPError(res)
	}
	return nil
}

// GetIssueURL returns the browsable (i.e. non-api) URL for the given issueID
func (c *Client) GetIssueURL(ctx context.Context, issueID string) (*url.URL, error) {
	path := fmt.Sprintf("issue/%s", issueID)
	u, err := c.resolvePath(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not resolve path: %s", path)
	}
	return u, nil
}

// GetIssue returns the Issue identified by the given issueID
func (c *Client) GetIssue(ctx context.Context, issueID string) (*Issue, error) {
	path := fmt.Sprintf("rest/issue/%s", issueID)
	req, err := c.newRequest(ctx, "GET", path)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request")
	}
	res, err := c.doWithLogin(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "error performing request")
	}
	defer res.Body.Close()
	if err := c.checkResponse(res); err != nil {
		return nil, errors.Wrap(err, "bad response")
	}
	issue := &Issue{}
	return issue, json.NewDecoder(res.Body).Decode(issue)
}
