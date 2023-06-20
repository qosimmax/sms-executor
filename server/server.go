// Package server provides functionality to easily set up a subscriber of pubsub events.
//
// The server holds all the clients it needs. The clients should be set up in the Create method.
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/qosimmax/sms-executor/client/redis"

	"github.com/qosimmax/sms-executor/client/smpp"

	"github.com/qosimmax/sms-executor/client/pubsub"
	"github.com/qosimmax/sms-executor/config"
	"github.com/qosimmax/sms-executor/monitoring/metrics"
	"github.com/qosimmax/sms-executor/monitoring/trace"
	"github.com/qosimmax/sms-executor/server/internal/event"
	"github.com/qosimmax/sms-executor/server/internal/handler"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// Server holds an HTTP server, config and all the clients.
type Server struct {
	Config  *config.Config
	HTTP    *http.Server
	PubSub  *pubsub.Client
	SMPP    *smpp.Client
	Storage *redis.Client
}

// Create sets up a server with necessary all clients.
// Returns an error if an error occurs.
func (s *Server) Create(ctx context.Context, config *config.Config) error {

	var psClient pubsub.Client
	if err := psClient.Init(ctx, config); err != nil {
		return fmt.Errorf("pubsub client: %w", err)
	}

	var rdClient redis.Client
	if err := rdClient.Init(ctx, config); err != nil {
		return fmt.Errorf("redis client: %w", err)
	}

	var smppClient smpp.Client
	if err := smppClient.Init(ctx, config); err != nil {
		return fmt.Errorf("smpp client: %w", err)
	}

	s.PubSub = &psClient
	s.SMPP = &smppClient
	s.Storage = &rdClient
	s.Config = config
	s.HTTP = &http.Server{
		Addr: fmt.Sprintf(":%s", s.Config.Port),
	}

	return nil
}

// Serve starts subscribing for messages.
// It also makes sure that the server gracefully shuts down on exit.
// Returns an error if an error occurs.
func (s *Server) Serve(ctx context.Context, errc chan<- error) {
	closer, err := trace.InitGlobalTracer(s.Config)

	if err != nil {
		errc <- err
	}

	defer closer.Close()

	go s.serveHTTP(errc)
	go s.subscribeAndListen(ctx, errc)

	log.Info("Ready")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Info("Shutdown signal received")

	s.shutdown(ctx)
}

func (s *Server) serveHTTP(errc chan<- error) {
	metrics.RegisterPrometheusCollectors() // modifies global state, yuck
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/_healthz", handler.Healthz)

	if err := s.HTTP.ListenAndServe(); err != http.ErrServerClosed {
		errc <- err
	}
}

func (s *Server) subscribeAndListen(ctx context.Context, errc chan<- error) {
	for _, e := range event.GetPubSubEvents(s.Storage, s.SMPP, s.Config) {
		go func(e event.PubSubEvent) {
			e.SubscribeAndListen(ctx, s.PubSub, errc)
		}(e)
	}
	for _, e := range event.GetAppEvents() {
		go func(e event.AppEvent) {
			e.SubscribeAndListen(ctx)
		}(e)
	}

	for _, e := range event.GetSmppEvents(s.PubSub, s.Storage) {
		go func(e event.SmppEvent) {
			e.SubscribeAndListen(ctx, s.SMPP)
		}(e)
	}

}

func (s *Server) shutdown(ctx context.Context) {
	if err := s.HTTP.Shutdown(ctx); err != nil {
		log.Error(err.Error())
	}

}
