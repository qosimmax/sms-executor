package event

import (
	"context"
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/opentracing/opentracing-go"
	"time"

	"github.com/qosimmax/sms-executor/client/pubsub"
	"github.com/qosimmax/sms-executor/monitoring/metrics"
	"github.com/qosimmax/sms-executor/monitoring/trace"
	"github.com/qosimmax/sms-executor/user"

	log "github.com/sirupsen/logrus"
)

// PubSubEvents contains a slice of PubSubEvent.
type PubSubEvents []PubSubEvent

// PubSubEvent contains the data for a PubSub event type.
type PubSubEvent struct {
	Name             string
	Queue            string
	SubscriptionName string
	Handler          Handler
	Subscription     nats.JetStreamContext
	Subscriptions    []Subscription
}

// SubscribeAndListen subscribes to a PubSubEvent.
func (e *PubSubEvent) SubscribeAndListen(ctx context.Context, c *pubsub.Client, errc chan<- error) {
	e.Subscription = c
	go e.receive(ctx, errc)
}

func (e *PubSubEvent) receive(ctx context.Context, errc chan<- error) {
	handler := func(ctx context.Context, msg *nats.Msg) {
		carrier := opentracing.TextMapCarrier{}
		for k := range msg.Header {
			carrier.Set(k, msg.Header.Get(k))
		}

		span, ctx := trace.ExtractFromCarrier(ctx, carrier, e.Name)
		defer span.Finish()

		metrics.ReceivedMessage(e.Name, float64(1))
		start := time.Now()
		defer func() {
			duration := time.Since(start)
			metrics.ObserveTimeToProcess(duration.Seconds())
		}()

		var errNonRecoverable user.ErrNonRecoverable
		var errExpected user.ErrExpected

		cctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		err := e.Handler.Handle(cctx, msg.Data)
		if err != nil {
			// If the error is not an expected error, log and record the error
			if !errors.As(err, &errExpected) {
				span.SetTag("error", true)
				span.LogKV("message", fmt.Errorf("error processing pubsub message: %w", err))
				log.Error(err.Error())
				metrics.OccurredError(e.Name)
			}

			// If the error is not a non-recoverable error, it means it is
			// recoverable, so return before acking
			if !errors.As(err, &errNonRecoverable) {
				return
			}
		}

		_ = msg.Ack()
	}

	for i, _ := range e.Subscriptions {
		sub, err := e.Subscription.PullSubscribe(e.Subscriptions[i].Name, e.Subscriptions[i].Queue,
			nats.DeliverAll(), nats.PullMaxWaiting(128))
		if err != nil {
			if err != nil {
				errc <- fmt.Errorf("subscription receive(%s): %w", e.Subscriptions[i].Name, err)
			}

			return
		}

		e.Subscriptions[i].sub = sub
		e.Subscriptions[i].id = i
	}

	// queue of subscriptions
	var queueSub *QueueSubscription
	for i := len(e.Subscriptions) - 1; i >= 0; i-- {
		push(&queueSub, e.Subscriptions[i])
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		for {
			msgs, err := queueSub.Sub.Fetch(queueSub.BatchSize, nats.MaxWait(queueSub.Timeout))
			if err == nats.ErrTimeout {
				break
			}

			for _, msg := range msgs {
				handler(ctx, msg)
			}

			// continue pull OTP type until timeout
			if queueSub.ID != 0 {
				break
			}

		}

		queueSub = queueSub.Next

	}

}

type QueueSubscription struct {
	ID        int
	Timeout   time.Duration
	BatchSize int
	Sub       *nats.Subscription
	Next      *QueueSubscription
}

type Subscription struct {
	Name      string
	Queue     string
	Timeout   time.Duration
	BatchSize int
	id        int
	sub       *nats.Subscription
}

func push(headRef **QueueSubscription, s Subscription) {
	ptr1 := &QueueSubscription{
		ID:        s.id,
		Timeout:   s.Timeout,
		BatchSize: s.BatchSize,
		Sub:       s.sub,
	}
	temp := *headRef
	ptr1.Next = *headRef

	if *headRef != nil {
		for temp.Next != *headRef {
			temp = temp.Next
		}

		temp.Next = ptr1
	} else {
		ptr1.Next = ptr1
	}

	*headRef = ptr1

}
