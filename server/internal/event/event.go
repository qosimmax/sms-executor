// Package event handles configuration and setup for receiving events.
//
// Events to subscribe to should be defined in GetPubSubEvents
package event

import (
	"context"
	"fmt"
	"time"

	"github.com/qosimmax/sms-executor/config"

	"github.com/qosimmax/sms-executor/client/smpp"

	"github.com/qosimmax/sms-executor/client/pubsub"

	"github.com/qosimmax/sms-executor/server/internal/handler"

	"github.com/qosimmax/sms-executor/client/redis"
)

// Handler is an interface that all event handles must implement.
type Handler interface {
	Handle(ctx context.Context, data []byte) error
}

// GetPubSubEvents describes all the pubsub events to listen to.
func GetPubSubEvents(r *redis.Client, s *smpp.Client, c *config.Config) PubSubEvents {
	psEvents := PubSubEvents{
		PubSubEvent{
			Name: "sms",
			Subscriptions: []Subscription{
				{
					Name:      fmt.Sprintf("sms.create.%s.otp", c.NatsTopic),
					Queue:     fmt.Sprintf("sms-executor:sms:otp:%s", c.NatsTopic),
					Timeout:   10 * time.Millisecond,
					BatchSize: c.RateLimit,
				},
				{
					Name:      fmt.Sprintf("sms.create.%s.default", c.NatsTopic),
					Queue:     fmt.Sprintf("sms-executor:sms:%s", c.NatsTopic),
					Timeout:   10 * time.Millisecond,
					BatchSize: c.RateLimit / 2,
				},
				{
					Name:      fmt.Sprintf("sms.create.%s.excel", c.NatsTopic),
					Queue:     fmt.Sprintf("sms-executor:sms:excel:%s", c.NatsTopic),
					Timeout:   10 * time.Millisecond,
					BatchSize: c.RateLimit / 2,
				},
			},
			Handler: &handler.Sms{
				SmsSender: s,
				Storage:   r,
			},
		},
	}

	return psEvents
}

// GetAppEvents describes all the app events to listen to.
func GetAppEvents() AppEvents {
	appEvents := AppEvents{}

	return appEvents
}

// GetSmppEvents describes all the smpp events to listen to.
func GetSmppEvents(ps *pubsub.Client, r *redis.Client) SmppEvents {
	smppEvents := SmppEvents{
		SmppEvent{
			Name: "SMPP",
			Handler: &handler.SmsEvent{
				Storage: r,
				Pub:     ps,
			},
		},
	}

	return smppEvents
}
