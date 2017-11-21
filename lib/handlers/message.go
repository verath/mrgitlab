package handlers

import (
	"context"

	"github.com/verath/mrgitlab/lib/gitlab"
)

// NewMessage return a new MergeRequestHandlerFunc that simply returns
// the provided message
func NewMessage(message string) MergeRequestHandlerFunc {
	return MergeRequestHandlerFunc(func(context.Context, *gitlab.MergeRequestWebhook) (string, error) {
		return message, nil
	})
}
