package main

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/qosimmax/sms-executor/config"
	"github.com/qosimmax/sms-executor/server"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	log.Info("Starting ...")

	ctx := context.Background()
	config, err := config.LoadConfig()

	if err != nil {
		log.Fatal(err.Error())
	}

	var s server.Server

	if err := s.Create(ctx, config); err != nil {
		log.Fatal(err.Error())
	}

	errc := make(chan error, 1)

	go func(errc chan error) {
		log.Fatal(<-errc)
	}(errc)

	s.Serve(ctx, errc)
}
