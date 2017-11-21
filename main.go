package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/verath/mrgitlab/lib"
	"github.com/verath/mrgitlab/lib/gitlab"
	"github.com/verath/mrgitlab/lib/handlers"
	"github.com/verath/mrgitlab/lib/youtrack"
)

func main() {
	serverAddr := flag.String("addr", ":3000",
		"TCP address where the server should listen for webhooks")
	gitlabBaseURL := flag.String("gitlab-url", "https://gitlab.com/",
		"The base URL of the GitLab server")
	gitlabToken := flag.String("gitlab-token", "",
		"The GitLab private token to use for request to the GitLab server")
	webhookToken := flag.String("webhook-token", "",
		"The webhook token that, if non-empty, must be included in the webhook calls")
	youtrackBaseURL := flag.String("youtrack-url", "http://track.example.com:8080/",
		"The base URL of the YouTrack server")
	youtrackUsername := flag.String("youtrack-username", "",
		"The YouTrack username of the user to use for authentication")
	youtrackPassword := flag.String("youtrack-password", "",
		"The YouTrack password of the user to use for authentication")
	debug := flag.Bool("debug", false,
		"Enables more verbose debug logging")
	flag.Parse()

	// Setup the logrus logger
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{DisableColors: true}
	if *debug {
		logger.Level = logrus.DebugLevel
		logger.Debug("Debug logging enabled")
	}

	// Setup the app, its dependencies, and register our handlers
	gitlabClient, err := gitlab.NewClient(logger, *gitlabBaseURL, *gitlabToken)
	if err != nil {
		logger.Fatalf("Error creating gitlabClient: %+v", err)
	}
	app, err := mrgitlab.New(logger, gitlabClient, *webhookToken)
	if err != nil {
		logger.Fatalf("Error creating app: %+v", err)
	}

	// Setup the YouTrack client
	youTrackClient, err := youtrack.NewClient(logger, *youtrackBaseURL,
		*youtrackUsername, *youtrackPassword)
	if err != nil {
		logger.Fatalf("error creating YouTrack Client: %+v", err)
	}

	// Register the merge request handlers. It is the handlers that provide
	// messages back to the gitlab merge request.
	youtrackMsg := handlers.NewYouTrack(youTrackClient, func(webhook *gitlab.MergeRequestWebhook) string {
		return youtrackBranchNameFilter(webhook.ObjectAttributes.SourceBranch)
	})
	beepBoopMsg := handlers.NewMessage("BeepBoop!")
	app.RegisterMergeRequestHandler("open", youtrackMsg)
	app.RegisterMergeRequestHandler("open", beepBoopMsg)

	// Setup an http server that forwards requests on "/" to the app
	// instance. We also define a /healthcheck endpoint for quick remote
	// health-checking.
	http.Handle("/", app)
	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	httpServer := &http.Server{Addr: *serverAddr}
	// Run the HTTP server, waiting for webhooks. We also
	// listen for interrupt signals (such as ctrl+c) to make
	// use stoppable.
	errCh := make(chan error)
	go func() { errCh <- httpServer.ListenAndServe() }()
	logger.Infof("HTTP server running at '%s'", httpServer.Addr)
	stopSigs := []os.Signal{os.Interrupt, os.Kill, syscall.SIGTERM}
	stopCh := make(chan os.Signal, len(stopSigs))
	signal.Notify(stopCh, stopSigs...)
	select {
	case err := <-errCh:
		logger.Fatalf("Error during ListenAndServe: %+v", err)
	case <-stopCh:
		logger.Info("Caught interrupt, shutting down...")
		httpServer.Close()
		<-errCh
	}
}

var youtrackBranchNameRegEx = regexp.MustCompile(`(?i)^(?:feature|release-fix)/([a-z]+)([0-9]+)`)

// youtrackBranchNameFilter takes a branch name and returns the YouTrack id
// associated with it, or an empty string if no id is associated with the
// branch name.
func youtrackBranchNameFilter(branchName string) string {
	matches := youtrackBranchNameRegEx.FindStringSubmatch(branchName)
	if matches != nil {
		return fmt.Sprintf("%s-%s", strings.ToUpper(matches[1]), matches[2])
	}
	return ""
}
