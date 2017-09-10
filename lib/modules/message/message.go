package message

import (
	"github.com/verath/mrgitlab/lib"
	"context"
	"github.com/verath/mrgitlab/lib/gitlab"
)

func New(message string) mrgitlab.MergeRequestHandler {
	return mrgitlab.MergeRequestHandlerFunc(func(context.Context, *gitlab.MergeRequestWebhook)(string, error) {
		return message, nil
	})
}
