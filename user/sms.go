package user

import (
	"context"
	"time"

	"github.com/qosimmax/gosmpp/data"
)

type SmsData struct {
	SmsID             string    `json:"sms_id"`
	Message           string    `json:"message"`
	Recipient         string    `json:"recipient"`
	CreatedAt         time.Time `json:"created_at"`
	NickName          string    `json:"nick_name"`
	TariffID          int       `json:"tariff_id"`
	CompanyID         string    `json:"company_id"`
	IsUnicode         bool      `json:"is_unicode"`
	SequenceNumber    int32     `json:"-"`
	SequenceMessageID string    `json:"-"`
}

func (s *SmsData) IsTimout() bool {
	if time.Now().Sub(s.CreatedAt).Seconds() > 10800 {
		return true
	}
	return false
}

func (s *SmsData) FindAndSetEncoding() {
	if data.FindEncoding(s.Message) == data.UCS2 {
		s.IsUnicode = true
	}
}

type SmsEvent struct {
	SmsID             string `json:"sms_id"`
	DestAddress       string `json:"destination_address"`
	SourceAddress     string `json:"source_address"`
	CommandStatus     string `json:"command_status"`
	SubmitDate        string `json:"submit_date"`
	DoneDate          string `json:"done_date"`
	DeliveryStatus    string `json:"delivery_status"`
	SequenceNumber    int32  `json:"sequence_number"`
	SequenceMessageID string `json:"sequence_message_id"`
	TariffID          int    `json:"tariff_id"`
	CompanyID         string `json:"company_id"`
	IsUnicode         bool   `json:"is_unicode"`
}

const (
	StatusExpired         = "SMS_EXPIRED"
	StatusConnectionError = "SMPP_CONN_ERROR"
	StatusSmsSent         = "SENT"
	StatusSmsDELIVERED    = "DELIVRD"
	StatusSmsFailed       = "FAILED"
)

// SmsSender is an interface for sending a sms
type SmsSender interface {
	SendSms(ctx context.Context, smsData SmsData) (err error)
}

// SequenceNumberReaderWriter is an interface for saving and getting a message sequence number
type SequenceNumberReaderWriter interface {
	WriteSequenceNumber(ctx context.Context, smsData SmsData) error
	ReadSequenceNumber(ctx context.Context, sequenceNumber int32) (SmsData, error)
}

// MessageSequenceReaderWriter is an interface for saving and getting a given message id
type MessageSequenceReaderWriter interface {
	WriteMessageSequence(ctx context.Context, smsData SmsData) error
	ReadMessageSequence(ctx context.Context, sequenceMessageID string) (SmsData, error)
}

type StorageReadWriter interface {
	SequenceNumberReaderWriter
	MessageSequenceReaderWriter
}

// SmsEventNotifier is an interface for notify other apps about sms statuses
type SmsEventNotifier interface {
	NotifySmsEvent(ctx context.Context, smsEvent SmsEvent) error
}
