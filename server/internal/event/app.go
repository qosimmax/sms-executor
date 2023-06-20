package event

import (
	"context"
	"errors"
	"github.com/opentracing/opentracing-go"
	"time"

	"github.com/qosimmax/sms-executor/user"

	log "github.com/sirupsen/logrus"
)

// AppEvents contains a slice of AppEvent.
type AppEvents []AppEvent

// AppEvent contains the data for an in-app event type.
type AppEvent struct {
	Name    string
	Rate    time.Duration
	Handler Handler
}

// SubscribeAndListen subscribes to an AppEvent.
func (e *AppEvent) SubscribeAndListen(ctx context.Context) {
	for t := range time.Tick(e.Rate) {
		go func(t time.Time) {
			span, ctx := opentracing.StartSpanFromContext(context.Background(), e.Name)
			defer span.Finish()

			var errExpected user.ErrExpected
			err := e.Handler.Handle(ctx, nil)
			if err != nil && !errors.As(err, &errExpected) {
				log.Error(t, err.Error())
			}
		}(t)
	}
}
