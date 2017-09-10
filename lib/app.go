package mrgitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/verath/mrgitlab/lib/gitlab"
	"net/http"
	"sync"
	"time"
)

// handleWebhookTimeout specifies the max amount of
// time that may elapse while handling a webhook
const handleWebhookTimeout = 2 * time.Minute

// App is the entry-point to the mrgitlab application. It implements
// the http handler interface for handling webhooks and should be registered
// to an http server.
type App struct {
	logger *logrus.Entry
	// webhookToken is the secret token we expect to receive from gitlab,
	// passed as the http header "X-Gitlab-Token"
	webhookToken string
	// gitlabClient is the rest client we use to send information
	// back to GitLab.
	gitlabClient *gitlab.Client

	mergeRequestHandlersMu sync.RWMutex
	// mergeRequestHandlers is a map from a merge request webhook action
	// (i.e. "open", "close", ...) to a slice of handlers for that action.
	mergeRequestHandlers map[string][]MergeRequestHandler
}

// New initializes an App instance. The webhookToken is a string that, if set, must also
// be present in all webhook calls.
func New(logger *logrus.Logger, gitlabClient *gitlab.Client, webhookToken string) (*App, error) {
	logEntry := logger.WithField("module", "mrgitlab")
	if webhookToken == "" {
		logEntry.Warn("No webhook token, all requests will be accepted!")
	}
	return &App{
		logger:               logEntry,
		gitlabClient:         gitlabClient,
		webhookToken:         webhookToken,
		mergeRequestHandlers: make(map[string][]MergeRequestHandler),
	}, nil
}

// RegisterMergeRequestHandler registers a MergeRequestHandler to the specified
// action. Action is the action specified by GitLab for the webhook. The following
// seems to be the only valid actions: "open", "close", "reopen", "merge".
func (app *App) RegisterMergeRequestHandler(action string, handler MergeRequestHandler) {
	app.mergeRequestHandlersMu.Lock()
	app.mergeRequestHandlers[action] = append(app.mergeRequestHandlers[action], handler)
	app.mergeRequestHandlersMu.Unlock()
}

// ServeHTTP is an http handler that is registered on the path that
// the GitLab webhook is posted to. It verifies and decodes the webhook
// from the http request.
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := app.ensureWebhookToken(r); err != nil {
		app.logger.Debugf("Bad webhook token: %+v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err := app.ensureMergeRequestWebhook(r); err != nil {
		app.logger.Debugf("Bad webhook event: %+v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	webhook := &gitlab.MergeRequestWebhook{}
	if err := json.NewDecoder(r.Body).Decode(webhook); err != nil {
		app.logger.Debugf("Error unmarshalling webhook: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), handleWebhookTimeout)
		defer cancel()
		if err := app.onMergeRequestWebhook(ctx, webhook); err != nil {
			app.logger.Debugf("Error handling webhook: %+v", err)
		}
	}()
}

// ensureWebhookToken checks the request for an "X-Gitlab-Token" and
// returns an error if it doesn't match the expected webhookToken.
func (app *App) ensureWebhookToken(r *http.Request) error {
	if app.webhookToken == "" {
		return nil
	}
	reqToken := r.Header.Get("X-Gitlab-Token")
	if reqToken != app.webhookToken {
		return errors.Errorf("X-Gitlab-Token missing or invalid, was: '%s'", reqToken)
	}
	return nil
}

// ensureMergeRequestWebhook tests the request for the "X-Gitlab-Event" header,
// returning an error if it does not exist or if it does not match the
// expected "Merge Request Hook" value.
func (app *App) ensureMergeRequestWebhook(r *http.Request) error {
	reqEvent := r.Header.Get("X-Gitlab-Event")
	if reqEvent != "Merge Request Hook" {
		return errors.Errorf("X-Gitlab-Event missing or invalid, was: '%s'", reqEvent)
	}
	return nil
}

// onMergeRequestWebhook is called when a merge requset webhook has been received
// and parsed successfully. onMergeRequestWebhook dispatches handling of the
// webhook to all registered MergeRequestWebhookHandler for the specific webhook
// action, waits for them to complete, then posts the accumulated message as a
// comment on the merge request.
func (app *App) onMergeRequestWebhook(ctx context.Context, webhook *gitlab.MergeRequestWebhook) error {
	app.logger.Debugf("onMergeRequestWebhook: %+v", webhook)
	action := webhook.ObjectAttributes.Action
	app.mergeRequestHandlersMu.RLock()
	handlers, ok := app.mergeRequestHandlers[action]
	app.mergeRequestHandlersMu.RUnlock()
	if !ok {
		app.logger.Debugf("No handler for Action: %s", action)
		return nil
	}
	// Fan-out, let each handler do its thing on a separate go-routine
	type handlerResult struct {
		msg string
		err error
	}
	var resultsChs []chan handlerResult
	for _, handler := range handlers {
		resultCh := make(chan handlerResult)
		go func(handler MergeRequestHandler) {
			msg, err := handler.HandleMergeRequest(ctx, webhook)
			resultCh <- handlerResult{msg, err}
		}(handler)
		resultsChs = append(resultsChs, resultCh)
	}
	// Fan-in, wait for each handler to complete (in order) and
	// combine their messages.
	var noteMessageBuf bytes.Buffer
	for _, resultCh := range resultsChs {
		res := <-resultCh
		if res.err != nil {
			app.logger.Debugf("Handler had an error in HandleMergeRequest: %+v", res.err)
			continue
		}
		if len(res.msg) > 0 {
			noteMessageBuf.WriteString(res.msg)
			noteMessageBuf.WriteByte('\n')
		}
	}
	// If no handler added anything to the message buffer, we send nothing.
	if noteMessageBuf.Len() == 0 {
		app.logger.Debugf("Not creating note, noteMessageBuf empty")
		return nil
	}
	mergeRequestID := gitlab.NewMergeRequestID(webhook)
	note := &gitlab.Note{Body: noteMessageBuf.String()}
	return errors.Wrap(app.gitlabClient.AddMergeRequestNote(ctx, mergeRequestID, note),
		"Error adding merge request note")
}
