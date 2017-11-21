package handlers

import (
	"context"
	"testing"

	"github.com/verath/mrgitlab/lib/gitlab"
)

func TestMessageHandler(t *testing.T) {
	testMsg := "teststr"
	h := NewMessage(testMsg)
	r, err := h.HandleMergeRequest(context.Background(), &gitlab.MergeRequestWebhook{})
	if err != nil {
		t.Errorf("did not expect an error, got: %+v", err)
	}
	if r != testMsg {
		t.Errorf("expected testMsg to equal '%s', was: '%s'", testMsg, r)
	}
}
