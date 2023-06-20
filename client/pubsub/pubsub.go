package pubsub

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/qosimmax/sms-executor/config"
	"github.com/qosimmax/sms-executor/monitoring/trace"
)

// Client holds the PubSub client.
type Client struct {
	nats.JetStreamContext
}

// Init sets up a new pubsub client.
func (c *Client) Init(ctx context.Context, config *config.Config) error {
	nc, err := nats.Connect(config.NatsURL)
	if err != nil {
		return err
	}

	js, err := nc.JetStream(nats.PublishAsyncMaxPending(10000))
	if err != nil {
		return err
	}

	_, err = js.StreamInfo("sms")
	if err != nil {
		if err == nats.ErrStreamNotFound {
			_, err = js.AddStream(&nats.StreamConfig{
				Name:     "sms",
				Subjects: []string{"sms.create.*.default", "sms.create.*.otp", "sms.create.*.excel"},
			})

			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	c.JetStreamContext = js
	return nil
}

func (c *Client) send(ctx context.Context, topicName string, data []byte) error {
	spanCarrier := trace.InjectIntoCarrier(ctx)

	msg := nats.NewMsg(topicName)
	msg.Data = data
	for k, v := range spanCarrier {
		msg.Header.Set(k, v)
	}

	_, err := c.PublishMsg(msg)
	if err != nil {
		return err
	}

	return nil
}
