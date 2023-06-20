package smpp

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"strconv"

	"github.com/qosimmax/gosmpp/data"
	"github.com/qosimmax/gosmpp/pdu"
	"github.com/qosimmax/sms-executor/user"
)

func (c *Client) SendSms(ctx context.Context, smsData user.SmsData) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "SendSms")
	defer span.Finish()

	submits, err := c.getMultiSubmitSM(smsData)
	if err != nil {
		return err
	}

	for i, _ := range submits {
		//ratelimit
		c.rl.Take()
		err = c.smpp.Transceiver().Submit(submits[i])
		if err != nil {
			return err
		}

	}

	return nil
}

func (c *Client) getMultiSubmitSM(smsData user.SmsData) (submits []*pdu.SubmitSM, err error) {
	enc := data.GSM7BIT
	if smsData.IsUnicode {
		enc = data.UCS2
	}

	var partitions []*pdu.ShortMessage
	if enc == data.UCS2 && len([]rune(smsData.Message)) <= 70 {
		var sm pdu.ShortMessage
		sm, err = pdu.NewShortMessageWithEncoding(smsData.Message, enc)
		partitions = append(partitions, &sm)
	} else {
		partitions, err = pdu.NewLongMessageWithEncoding(smsData.Message, enc)
	}

	if err != nil {
		return nil, fmt.Errorf("error get message partitions:%w", err)
	}

	isNumeric, _ := strconv.Atoi(smsData.NickName)
	srcAddr := pdu.NewAddress()
	if isNumeric > 0 {
		srcAddr.SetTon(3)
	} else {
		srcAddr.SetTon(5)
	}

	srcAddr.SetNpi(0)

	err = srcAddr.SetAddress(smsData.NickName)
	if err != nil {
		return nil, fmt.Errorf("error set source address in smpp:%w", err)
	}

	destAddr := pdu.NewAddress()
	destAddr.SetTon(1)
	destAddr.SetNpi(1)
	err = destAddr.SetAddress(smsData.Recipient)
	if err != nil {
		return nil, fmt.Errorf("error set destination address in smpp:%w", err)
	}

	for i, _ := range partitions {
		submitSM := pdu.NewSubmitSM().(*pdu.SubmitSM)
		submitSM.SourceAddr = srcAddr
		submitSM.DestAddr = destAddr
		submitSM.ProtocolID = 0
		submitSM.RegisteredDelivery = 1
		submitSM.EsmClass = 0
		submitSM.ReplaceIfPresentFlag = 0
		submitSM.SequenceNumber = smsData.SequenceNumber
		submitSM.Message = *partitions[i]
		if len(partitions) > 1 {
			submitSM.EsmClass = 0x40
		}

		submits = append(submits, submitSM)
	}

	return
}

func (c *Client) Events(ctx context.Context) chan user.SmsEvent {
	return c.events
}
