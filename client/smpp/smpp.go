package smpp

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"go.uber.org/ratelimit"

	"github.com/qosimmax/gosmpp/data"

	"github.com/qosimmax/sms-executor/user"

	"github.com/qosimmax/gosmpp/pdu"

	"github.com/qosimmax/gosmpp"
	"github.com/qosimmax/sms-executor/config"
)

// Client holds the SMPP client.
type Client struct {
	smpp         *gosmpp.Session
	events       chan user.SmsEvent
	rl           ratelimit.Limiter
	operatorName string
}

func (c *Client) Init(ctx context.Context, config *config.Config) (err error) {
	c.events = make(chan user.SmsEvent, 100)
	c.rl = ratelimit.New(config.RateLimit)
	auth := gosmpp.Auth{
		SMSC:       config.OperatorURL,
		SystemID:   config.OperatorLogin,
		Password:   config.OperatorPassword,
		SystemType: "",
	}

	c.operatorName = config.NatsTopic

	c.smpp, err = gosmpp.NewSession(
		gosmpp.TRXConnector(gosmpp.NonTLSDialer, auth),
		gosmpp.Settings{
			EnquireLink: 5 * time.Second,

			ReadTimeout: 10 * time.Second,

			OnSubmitError: func(_ pdu.PDU, err error) {
				log.Println("SubmitPDU error:", err)
			},

			OnReceivingError: func(err error) {
				log.Println("Receiving PDU/Network error:", err)
			},

			OnRebindingError: func(err error) {
				log.Println("Rebinding but error:", err)
			},

			OnPDU: c.handlePDU(),

			OnClosed: func(state gosmpp.State) {
				log.Println(state)
			},
		}, 5*time.Second)

	if err != nil {
		return fmt.Errorf("error connect smpp server:%w, host=%s, login=%s, pass=%s", err,
			auth.SMSC, auth.SystemID, auth.Password)
	}

	return nil
}

func (c *Client) handlePDU() func(pdu.PDU, bool) {
	return func(p pdu.PDU, _ bool) {
		switch pd := p.(type) {
		case *pdu.SubmitSMResp:
			deliveryStatus := user.StatusSmsSent
			if pd.CommandStatus != data.ESME_ROK {
				deliveryStatus = user.StatusSmsFailed
			}

			c.events <- user.SmsEvent{
				SequenceMessageID: pd.MessageID,
				CommandStatus:     pd.CommandStatus.String(),
				SequenceNumber:    pd.SequenceNumber,
				DeliveryStatus:    deliveryStatus,
				SubmitDate:        time.Now().Format(time.RFC3339),
				DoneDate:          time.Now().Format(time.RFC3339),
			}

		case *pdu.GenericNack:
			log.Println("GenericNack Received")

		case *pdu.EnquireLinkResp:
			log.Println("EnquireLinkResp Received")

		case *pdu.DataSM:
			log.Printf("DataSM:%+v\n", pd)

		case *pdu.DeliverSM:
			//log.Printf("DeliverSM:%+v\n", pd)

			message, _ := pd.Message.GetMessage()
			values, _ := parseMessage(message)

			messageId := strings.TrimPrefix(values["id"], "0")
			messageId = strings.TrimPrefix(messageId, "0")

			c.events <- user.SmsEvent{
				SequenceMessageID: messageId,
				DestAddress:       pd.SourceAddr.Address(),
				SourceAddress:     pd.DestAddr.Address(),
				CommandStatus:     pd.CommandStatus.String(),
				SubmitDate:        time.Now().Format(time.RFC3339),
				DoneDate:          time.Now().Format(time.RFC3339),
				//SubmitDate:        values["submit_date"],
				//DoneDate:          values["done_date"],
				DeliveryStatus: values["stat"],
				SequenceNumber: pd.SequenceNumber,
			}

		}
	}
}

// Regex pattern captures "key: value" pair from the content.
var pattern = regexp.MustCompile(`(?m)(?P<key>\w+):(?P<value>\w+)`)

func parseMessage(message string) (map[string]string, error) {

	values := make(map[string]string)
	content := strings.Replace(message, "done date", "done_date", 1)
	content = strings.Replace(content, "submit date", "submit_date", 1)

	for _, sub := range pattern.FindAllStringSubmatch(content, -1) {
		values[sub[1]] = sub[2]
	}

	t1, err := time.Parse("0601021504", values["submit_date"])
	if err != nil {
		return nil, err
	}
	t2, err := time.Parse("0601021504", values["done_date"])
	if err != nil {
		return nil, err
	}

	//@TODO fix time zone
	values["submit_date"] = t1.Add(-5 * time.Hour).Format(time.RFC3339)
	values["done_date"] = t2.Add(-5 * time.Hour).Format(time.RFC3339)

	return values, nil

}
