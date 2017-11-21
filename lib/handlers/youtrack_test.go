package handlers

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/pkg/errors"
	"github.com/verath/mrgitlab/lib/gitlab"
	"github.com/verath/mrgitlab/lib/youtrack"
)

var issueWithoutDescJSON = []byte(`{
	"field": [
		{
            "name": "summary",
            "value": "Product X example code problems"
        }
	]
}`)

// mock implementation of the youTrackClient interface.
type mockYouTrackClient struct {
	GetIssueFunc    func(ctx context.Context, issueID string) (*youtrack.Issue, error)
	GetIssueURLFunc func(ctx context.Context, issueID string) (*url.URL, error)
}

func (c *mockYouTrackClient) GetIssue(ctx context.Context, issueID string) (*youtrack.Issue, error) {
	return c.GetIssueFunc(ctx, issueID)
}

func (c *mockYouTrackClient) GetIssueURL(ctx context.Context, issueID string) (*url.URL, error) {
	return c.GetIssueURLFunc(ctx, issueID)
}

// Test that we don't send a message if the branch filter
// function returns an empty issueID
func TestYouTrackHandler_NoBranchID(t *testing.T) {
	mockClient := &mockYouTrackClient{}
	filterFunc := func(*gitlab.MergeRequestWebhook) string {
		return ""
	}
	h := NewYouTrack(mockClient, filterFunc)
	msg, err := h.HandleMergeRequest(context.Background(), nil)
	if err != nil {
		t.Fatalf("Unexpected error handling merge request: %+v", err)
	}
	if msg != "" {
		t.Errorf("Expected msg to be empty, was '%s'", msg)
	}
}

func TestYouTrackHandler_NoIssueURL(t *testing.T) {
	mockClient := &mockYouTrackClient{}
	mockClient.GetIssueURLFunc = func(context.Context, string) (*url.URL, error) {
		return nil, errors.New("testerr")
	}
	filterFunc := func(*gitlab.MergeRequestWebhook) string {
		return "ISSUEID"
	}
	h := NewYouTrack(mockClient, filterFunc)
	msg, err := h.HandleMergeRequest(context.Background(), nil)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
	if msg != "" {
		t.Errorf("Expected msg to be empty, was '%s'", msg)
	}
}

func TestYouTrackHandler_FailFetchingIssue(t *testing.T) {
	mockClient := &mockYouTrackClient{}
	mockClient.GetIssueURLFunc = func(context.Context, string) (*url.URL, error) {
		return url.Parse("http://youtrack.test")
	}
	mockClient.GetIssueFunc = func(context.Context, string) (*youtrack.Issue, error) {
		return nil, errors.New("testerr")
	}
	filterFunc := func(*gitlab.MergeRequestWebhook) string {
		return "ISSUEID"
	}
	h := NewYouTrack(mockClient, filterFunc)
	msg, err := h.HandleMergeRequest(context.Background(), nil)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
	if msg != "" {
		t.Errorf("Expected msg to be empty, was '%s'", msg)
	}
}

// Test that we don't return an error only because the description
// field of the issue is empty.
func TestYouTrackHandler_EmptyDescription(t *testing.T) {
	issueWithoutDesc := &youtrack.Issue{}
	if err := json.Unmarshal(issueWithoutDescJSON, issueWithoutDesc); err != nil {
		panic(err) // json decode not part of what we test
	}
	mockClient := &mockYouTrackClient{}
	mockClient.GetIssueURLFunc = func(context.Context, string) (*url.URL, error) {
		return url.Parse("http://youtrack.test")
	}
	mockClient.GetIssueFunc = func(context.Context, string) (*youtrack.Issue, error) {
		return issueWithoutDesc, nil
	}
	filterFunc := func(*gitlab.MergeRequestWebhook) string {
		return "ISSUEID"
	}
	h := NewYouTrack(mockClient, filterFunc)
	_, err := h.HandleMergeRequest(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error, but got: %+v", err)
	}
}
