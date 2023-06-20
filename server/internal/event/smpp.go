package event

import (
	"context"
	"encoding/json"
	"github.com/opentracing/opentracing-go"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/qosimmax/sms-executor/client/smpp"
)

// SmppEvents contains a slice of SmppEvent.
type SmppEvents []SmppEvent

// SmppEvent contains the data for a smpp event type.
type SmppEvent struct {
	Name    string
	Rate    time.Duration
	Handler Handler
}

// SubscribeAndListen subscribes to an AppEvent.
func (e *SmppEvent) SubscribeAndListen(ctx context.Context, c *smpp.Client) {
	events := c.Events(ctx)
	for {
		select {
		case event := <-events:
			func() {
				span, ctx := opentracing.StartSpanFromContext(context.Background(), e.Name)
				defer span.Finish()

				data, _ := json.Marshal(event)
				err := e.Handler.Handle(ctx, data)
				if err != nil {
					log.Errorf("error on smpp event: %v", err)
				}
				span.SetTag("event", true)
				span.LogKV("event", event)
			}()

		case <-ctx.Done():
			log.Info("done")
			return
		}
	}

}
