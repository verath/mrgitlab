package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/verath/mrgitlab/lib/gitlab"
	"github.com/verath/mrgitlab/lib/youtrack"
)

// youTrackClient is an interface abstracting the YouTrack client used for API
// calls. This is primarily so that we can do unit tests against a non-network
// implementation.
type youTrackClient interface {
	GetIssue(ctx context.Context, issueID string) (*youtrack.Issue, error)
	GetIssueURL(ctx context.Context, issueID string) (*url.URL, error)
}

// youtrackWebhookFilterFunc is a function that, given a webhook, returns the YouTrack
// issue id associated with it. If no id could be extracted from the webhook, the
// function should return an empty string.
type youtrackWebhookFilterFunc func(webhook *gitlab.MergeRequestWebhook) (issueID string)

// NewYouTrack creates a new MergeRequestHandlerFunc that uses the provided YouTrackClient
// to lookup the YouTrack issue associated with a merge request and adds the issue
// data to the merge request comment. The filter parameter specifies a filter that is
// used to discard merge requests that are not associated with YouTrack.
func NewYouTrack(client youTrackClient, filter youtrackWebhookFilterFunc) MergeRequestHandlerFunc {
	if client == nil || filter == nil {
		panic("client and filter must not be nil")
	}
	return MergeRequestHandlerFunc(func(ctx context.Context, webhook *gitlab.MergeRequestWebhook) (string, error) {
		issueID := filter(webhook)
		if issueID == "" {
			return "", nil
		}
		issueURL, err := client.GetIssueURL(ctx, issueID)
		if err != nil {
			return "", errors.Wrapf(err, "could not resolve issue URL for issueID '%s'", issueID)
		}
		issue, err := client.GetIssue(ctx, issueID)
		if err != nil {
			if youtrack.IsHTTPStatusError(err, http.StatusNotFound) {
				// We don't treat not found as an error as it could just
				// be that the branch name just looked like a youtrack id.
				return "", nil
			}
			return "", errors.Wrapf(err, "could not get issue for issueID '%s'", issueID)
		}
		issueSummary, err := issue.FieldStringValue("summary")
		if err != nil {
			return "", errors.Wrap(err, "could not get value for 'summary'")
		}
		issueDescription, err := issue.FieldStringValue("description")
		if err != nil {
			// We don't see the description missing as an error as
			// description is not a mandatory field in YouTrack.
			issueDescription = ""
		}
		// We filter special GitLab reference here, so that we don't accidentally
		// spam users by mentioning them in the comment
		issueSummary = filterGitLabReferences(issueSummary)
		issueDescription = filterGitLabReferences(issueDescription)
		issueDescription = markdownQuote(issueDescription)
		issueTitle := issueID + ": " + issueSummary
		return fmt.Sprintf(""+
			"# %s\n"+
			"%s\n\n"+
			"%s\n",
			issueTitle, issueURL, issueDescription), nil
	})
}
