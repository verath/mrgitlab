package youtrack

import (
	"context"
	"net/url"
	"testing"

	"github.com/Sirupsen/logrus"
)

// Test that the login ticket channel is initialized with (at least) one ticket.
// Without a ticket added, no request would ever be possible as they would
// always time out waiting for a new ticket
func TestNewClient_AddsLoginTicket(t *testing.T) {
	c, err := NewClient(logrus.New(), "http://track.example.com:8080/", "user", "pass")
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	select {
	case <-c.loginTicketCh:
	default:
		t.Error("loginTicketCh was blocking (no ticket added)")
	}
}

func TestGetIssueURL(t *testing.T) {
	c, err := NewClient(logrus.New(), "http://track.example.com:8080/", "user", "pass")
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	expected, _ := url.Parse("http://track.example.com:8080/issue/XYZ-989")
	actual, err := c.GetIssueURL(context.Background(), "XYZ-989")
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	if actual.String() != expected.String() {
		t.Errorf("expected '%s', got: '%s'", expected, actual)
	}
}
