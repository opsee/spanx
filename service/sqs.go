package service

import (
	"encoding/json"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	log "github.com/sirupsen/logrus"
)

const (
	pollerMaxMessageCount = 1
)

// A Poller is a generic SQS poller that does 20-second long-polling in a
// separate goroutine.
type Poller struct {
	client   sqsiface.SQSAPI
	queueURL string
	handler  MessageHandler
	msgChan  chan interface{}
	stopChan chan interface{}
	stopped  chan interface{}
	stopping bool
}

// MessageHandler is a function that takes a string message body and either
// takes action on a message or returns an error (causing the message to
// eventually be requeued.)
type MessageHandler func(string) error

// NewPoller creates and starts a poller for the specified SQS queue at `url`.
func NewPoller(sqs sqsiface.SQSAPI, url string, handler MessageHandler) (*Poller, error) {
	p := Poller{
		client:   sqs,
		queueURL: url,
		handler:  handler,
	}
	p.msgChan = make(chan interface{}, pollerMaxMessageCount)
	p.stopChan = make(chan interface{}, 1)
	p.stopping = false
	go p.poll()
	return p
}

func (p *Poller) poll() {
	for {
		resp, err := p.client.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            p.QueueURL,
			AttributeNames:      []*string{aws.String("All")},
			MaxNumberOfMessages: aws.Int64(pollerMaxMessageCount),
			WaitTimeSeconds:     aws.Int64(20),
		})
		if err != nil {
			log.WithError(err).Error("Error calling SQS ReceiveMessage")
		}

		entries := make([]*sqs.DeleteMessageBatchRequestEntry, pollerMaxMessageCount)

		for i, msg := range resp.Messages {
			senderID := aws.StringValue(msg.Attributes["SenderId"])
			messageID := aws.StringValue(msg.MessageId)
			sentTimestamp := aws.StringValue(msg.Attributes["SentTimestamp"])
			i, err := strconv.ParseInt(sentTimestamp)
			if err != nil {
				log.WithError(err).Errorf("Unable to parse sent timestamp from SQS message: %s", sentTimestamp)
			}

			now := time.Now()
			sent := time.Unix(sentTimestamp, 0)
			delay := now.Sub(sent)

			log.WithFields(log.Fields{
				"message_id": messageID,
				"sender_id":  senderID,
				"sent":       sent.String(),
				"delay_ms":   int64(delay / time.Millisecond),
				"message":    msg,
			}).Info("Received message from SQS.")

			if err := p.Handler(aws.StringValue(msg.Body)); err != nil {
				log.WithFields(log.Fields{
					"message_id": messageID,
					"sender_id":  senderID,
				}).WithError(err).Error("Unable to handle message")
				continue
			}

			entries[i] = &sqs.DeleteMessageBatchRequestEntry{
				ReceiptHandle: msg.ReceiptHandle,
				Id:            msg.Id,
			}
		}

		resp, err = sqs.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
			QueueUrl: p.queueURL,
			Entries:  entries,
		})
		if err != nil {
			log.WithError(err).Error("Error calling SQS DeleteMessageBatch")
		}
	}
}
