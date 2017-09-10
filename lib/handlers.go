package mrgitlab

import (
	"context"
	"github.com/verath/mrgitlab/lib/gitlab"
)

// MergeRequestHandler is a handler for handling merge requests
// events, triggered via GitLab webhooks. The assumption for a
// MergeRequestHandler is that it performs some action and then
// wants to add a comment back to the merge request providing some
// additional context.
type MergeRequestHandler interface {
	// HandleMergeRequest is called when a merge request webhook have
	// been received. The HandleMergeRequest must return the context's
	// error should the context become cancelled before the handler
	// can finish. The HandleMergeRequest must not modify the provided
	// MergeRequestWebhook data. On success, a string may be returned
	// which will be appended to a new note added to the merge request.
	HandleMergeRequest(context.Context, *gitlab.MergeRequestWebhook) (string, error)
}

// MergeRequestHandlerFunc is a wrapper allowing a func
// to implement the MergeRequestHandler interface
type MergeRequestHandlerFunc func(context.Context, *gitlab.MergeRequestWebhook) (string, error)

func (f MergeRequestHandlerFunc) HandleMergeRequest(ctx context.Context, webhook *gitlab.MergeRequestWebhook) (string, error) {
	return f(ctx, webhook)
}
