package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/opentracing/opentracing-go"

	"github.com/qosimmax/sms-executor/user"
)

func (c *Client) NotifySmsEvent(ctx context.Context, smsEvent user.SmsEvent) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NotifySmsEvent")
	defer span.Finish()

	data, err := json.Marshal(smsEvent)
	if err != nil {
		return fmt.Errorf("error marshalling sms-event data to send to pubsub: %w", err)
	}

	err = c.send(ctx, fmt.Sprintf("sms.events.%s", smsEvent.DeliveryStatus), data)
	if err != nil {
		return fmt.Errorf("error sending sms-event data message to pubsub: %w", err)
	}

	return nil
}
