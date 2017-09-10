package main

import (
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/verath/mrgitlab/lib"
	"github.com/verath/mrgitlab/lib/gitlab"
	"github.com/verath/mrgitlab/lib/modules/message"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	debug := flag.Bool("debug", false,
		"Enables more verbose debug logging")
	flag.Parse()

	// Setup the logrus logger
	logger := logrus.New()
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

	// Register the merge request handlers. It is the handlers that provide
	// messages back to the gitlab merge request.
	beepBoopMsg := message.New(":robot: Beep Boop!")
	app.RegisterMergeRequestHandler("open", beepBoopMsg)
	app.RegisterMergeRequestHandler("reopen", beepBoopMsg)

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
