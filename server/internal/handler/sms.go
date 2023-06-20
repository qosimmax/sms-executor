package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/qosimmax/sms-executor/user"
)

type Sms struct {
	SmsSender      user.SmsSender
	Storage        user.StorageReadWriter
	sequenceNumber int32
}

func (s *Sms) Handle(ctx context.Context, data []byte) error {
	var smsData user.SmsData
	err := json.Unmarshal(data, &smsData)
	if err != nil {
		return user.ErrNonRecoverable{
			Err: fmt.Errorf("failed to unmarshal sms data in sms handle: %w", err),
		}
	}

	if smsData.IsTimout() {
		return user.ErrNonRecoverable{
			Err: fmt.Errorf("sms message live timeout"),
		}
	}

	// set message sequence number
	smsData.SequenceNumber = s.incSeqNumber()
	smsData.FindAndSetEncoding()
	err = s.Storage.WriteSequenceNumber(ctx, smsData)
	if err != nil {
		return fmt.Errorf("error on write sequenceNumber in sms handle: %w", err)
	}

	err = s.SmsSender.SendSms(ctx, smsData)
	if err != nil {
		return fmt.Errorf("error sending sms message in sms handle: %w", err)
	}

	return nil
}

type SmsEvent struct {
	Storage user.StorageReadWriter
	Pub     user.SmsEventNotifier
}

func (s *SmsEvent) Handle(ctx context.Context, data []byte) error {
	var smsEvent user.SmsEvent
	err := json.Unmarshal(data, &smsEvent)
	if err != nil {
		return user.ErrNonRecoverable{
			Err: fmt.Errorf("failed to unmarshal sms event data in sms event handler: %w", err),
		}
	}

	switch smsEvent.DeliveryStatus {
	case user.StatusSmsSent:
		seqNum, err := s.Storage.ReadSequenceNumber(ctx, smsEvent.SequenceNumber)
		if err != nil {
			return err
		}
		seqNum.SequenceMessageID = smsEvent.SequenceMessageID
		smsEvent.SmsID = seqNum.SmsID
		smsEvent.DestAddress = seqNum.Recipient
		smsEvent.CompanyID = seqNum.CompanyID
		smsEvent.TariffID = seqNum.TariffID
		smsEvent.IsUnicode = seqNum.IsUnicode

		err = s.Storage.WriteMessageSequence(ctx, seqNum)
		if err != nil {
			return err
		}
	case user.StatusSmsFailed:
		seqNum, err := s.Storage.ReadSequenceNumber(ctx, smsEvent.SequenceNumber)
		if err != nil {
			return err
		}

		smsEvent.SmsID = seqNum.SmsID
		smsEvent.DestAddress = seqNum.Recipient
		smsEvent.CompanyID = seqNum.CompanyID
		smsEvent.TariffID = seqNum.TariffID
		smsEvent.IsUnicode = seqNum.IsUnicode

	default: //DELIVERED, UNDELIVERED, REJECTED, EXPIRED..., etc.
		seqMsg, err := s.Storage.ReadMessageSequence(ctx, smsEvent.SequenceMessageID)
		if err != nil {
			return err
		}

		smsEvent.SmsID = seqMsg.SmsID
		smsEvent.CompanyID = seqMsg.CompanyID
		smsEvent.TariffID = seqMsg.TariffID
		smsEvent.IsUnicode = seqMsg.IsUnicode
	}

	log.Println("sms event", smsEvent)
	if smsEvent.SmsID == "" {
		return user.ErrNonRecoverable{
			Err: fmt.Errorf("sms_id is empty"),
		}
	}

	err = s.Pub.NotifySmsEvent(ctx, smsEvent)
	if err != nil {
		return err
	}

	return nil
}

func (c *Sms) incSeqNumber() int32 {
	if c.sequenceNumber >= math.MaxInt32-1 {
		c.sequenceNumber = 0
	}
	c.sequenceNumber += 1

	return c.sequenceNumber
}
