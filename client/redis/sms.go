package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"

	"github.com/qosimmax/sms-executor/user"
)

func (c *Client) WriteSequenceNumber(ctx context.Context, smsData user.SmsData) error {
	key := fmt.Sprintf("seqId:%s:%d", c.topic, smsData.SequenceNumber)
	smsData.Message = ""
	data, _ := json.Marshal(smsData)
	err := c.redis.Set(ctx, key, data, 900*time.Second).Err()
	return err
}

func (c *Client) ReadSequenceNumber(ctx context.Context, sequenceNumber int32) (smsData user.SmsData, err error) {
	key := fmt.Sprintf("seqId:%s:%d", c.topic, sequenceNumber)
	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return user.SmsData{}, nil
		}
		return user.SmsData{}, err
	}

	err = json.Unmarshal(data, &smsData)
	return
}

func (c *Client) WriteMessageSequence(ctx context.Context, smsData user.SmsData) error {
	key := fmt.Sprintf("seqMsgID:%s:%s", c.topic, smsData.SequenceMessageID)
	data, _ := json.Marshal(smsData)
	err := c.redis.Set(ctx, key, data, 24*time.Hour+time.Minute).Err()
	return err
}

func (c *Client) ReadMessageSequence(ctx context.Context, sequenceMessageID string) (smsData user.SmsData, err error) {
	key := fmt.Sprintf("seqMsgID:%s:%s", c.topic, sequenceMessageID)
	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return user.SmsData{}, nil
		}
		return user.SmsData{}, err
	}

	_ = c.redis.Del(ctx, key).Err()
	err = json.Unmarshal(data, &smsData)
	return
}
