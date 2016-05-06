package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/opsee/basic/com"
	"github.com/opsee/spanx/store"
	log "github.com/sirupsen/logrus"
)

const (
	CreateSuccess = "SUCCESS"
	CreateFailed  = "FAILED"
)

var (
	httpClient = &http.Client{}
)

type CFNotification struct {
	Type             string `json:"Type"`
	MessageId        string `json:"MessageId"`
	TopicArn         string `json:"TopicArn"`
	Subject          string `json:"Subject"`
	Message          string `json:"Message"`
	Timestamp        string `json:"Timestamp"`
	SignatureVersion string `json:"SignatureVersion"`
	Signature        string `json:"Signature"`
	SigningCertURL   string `json:"SigningCertURL"`
	UnsubscribeURL   string `json:"UnsubscribeURL"`
}

type CallbackMessage struct {
	RequestType        string                     `json:"RequestType"`
	ServiceToken       string                     `json:"ServiceToken"`
	ResponseURL        string                     `json:"ResponseURL"`
	StackId            string                     `json:"StackId"`
	RequestId          string                     `json:"RequestId"`
	LogicalResourceId  string                     `json:"LogicalResourceId"`
	PhysicalResourceId string                     `json:"PhysicalResourceId"`
	ResourceType       string                     `json:"ResourceType"`
	ResourceProperties CallbackResourceProperties `json:"ResourceProperties"`
}

type CallbackResourceProperties struct {
	ServiceToken   string `json:"ServiceToken"`
	RoleExternalID string `json:"RoleExternalID"`
	RoleARN        string `json:"RoleARN"`
	StackName      string `json:"StackName"`
	StackID        string `json:"StackID"`
	Region         string `json:"Region"`
}

type CallbackResponse struct {
	Status             string            `json:"Status"`
	Reason             string            `json:"Reason,omitempty"`
	LogicalResourceId  string            `json:"LogicalResourceId"`
	RequestId          string            `json:"RequestId"`
	StackId            string            `json:"StackId"`
	PhysicalResourceId string            `json:"PhysicalResourceId,omitempty"`
	Data               map[string]string `json:"Data,omitempty"`
}

func (s *service) putCallbackResponse(response interface{}, url string) error {
	body, err := json.Marshal(response)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, strings.NewReader(string(body)))
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Received bad response code from S3: %i", resp.StatusCode)
	}

	return nil
}

func makeResponse(cbk CallbackMessage, success bool) *CallbackResponse {
	response := &CallbackResponse{
		LogicalResourceId: cbk.LogicalResourceId,
		RequestId:         cbk.RequestId,
		StackId:           cbk.StackId,
	}

	if success {
		response.Status = "SUCCESS"
		if cbk.PhysicalResourceId != "" {
			response.PhysicalResourceId = cbk.PhysicalResourceId
		} else {
			response.PhysicalResourceId = cbk.ResourceProperties.RoleExternalID
		}
	} else {
		response.Status = "FAILURE"
		// I don't want to leak any information to an external party
		// about why the request failed. All of the specific failure
		// types are logged on our end.
		response.Reason = "Error processing request."
	}

	return response

}

func (s *service) handleCFCreateRequest(cbk CallbackMessage) error {
	var (
		stack      *store.Stack
		account    *com.Account
		customerID string
		err        error
	)

	externalID := cbk.ResourceProperties.RoleExternalID
	if externalID == "" {
		log.WithFields(log.Fields{
			"request_id": cbk.RequestId,
		}).Error("Callback does not contain an external ID.")
		return s.putCallbackResponse(makeResponse(cbk, false), cbk.ResponseURL)
	}

	// 1. Get the Account object from the DB
	account, err = s.db.GetAccountByExternalID(externalID)
	if err != nil {
		log.WithFields(log.Fields{
			"request_id": cbk.RequestId,
		}).Error("Error getting external ID")
		return s.putCallbackResponse(makeResponse(cbk, false), cbk.ResponseURL)
	}

	if account == nil {
		log.WithFields(log.Fields{
			"request_id": cbk.RequestId,
		}).Error("External ID not found.")
		return s.putCallbackResponse(makeResponse(cbk, false), cbk.ResponseURL)
	}

	customerID = account.CustomerID

	stack, err = s.db.GetStack(customerID, externalID)
	if err != nil {
		// This could be a transient error. For now, just return an error
		// so that the message is requeued. If it's a permanent failure,
		// then it will end up in the dead letter queue, and we can
		// investigate it.
		//
		// TODO(greg): Add monitoring of the dead letter queue (opsee-cfn-deadletter)
		return err
	}

	if stack != nil {
		log.WithFields(log.Fields{
			"request_id":  cbk.RequestId,
			"customer_id": customerID,
			"external_id": externalID,
		}).Info("Handling create for existing stack.")

		newStack := *stack
		newStack.StackID = cbk.StackId
		newStack.StackName = cbk.ResourceProperties.StackName
		newStack.Active = true
		newStack.Region = cbk.ResourceProperties.Region
		err = s.db.UpdateStack(stack, &newStack)
		if err != nil {
			return err
		}
	} else {
		if cbk.StackId == "" {
			log.WithFields(log.Fields{
				"customer_id": customerID,
				"external_id": externalID,
				"request_id":  cbk.RequestId,
			}).WithError(err).Error("Callback contained no stack ID.")
			return s.putCallbackResponse(makeResponse(cbk, false), cbk.ResponseURL)
		}

		if cbk.ResourceProperties.StackName == "" {
			log.WithFields(log.Fields{
				"customer_id": customerID,
				"external_id": externalID,
				"request_id":  cbk.RequestId,
			}).WithError(err).Error("Callback contained no stack name.")
			return s.putCallbackResponse(makeResponse(cbk, false), cbk.ResponseURL)
		}

		log.WithFields(log.Fields{
			"customer_id": customerID,
			"external_id": externalID,
			"stack_id":    cbk.StackId,
		}).Info("Handling create for new stack.")

		stack = &store.Stack{
			StackID:    cbk.StackId,
			StackName:  cbk.ResourceProperties.StackName,
			CustomerID: customerID,
			ExternalID: externalID,
			Region:     cbk.ResourceProperties.Region,
			Active:     true,
		}

		err = s.db.PutStack(stack)
		if err != nil {
			return err
		}
	}

	account.Active = true
	account.RoleARN = cbk.ResourceProperties.RoleARN
	err = s.db.UpdateAccount(account)
	if err != nil {
		return err
	}

	return s.putCallbackResponse(makeResponse(cbk, true), cbk.ResponseURL)
}

// We will eventually want some kind of role policy version for feature
// gating. We can handle updating our internal policy version with this
// method.
func (s *service) handleCFUpdateRequest(cbk CallbackMessage) error {
	externalID := cbk.ResourceProperties.RoleExternalID
	if externalID == "" {
		log.WithFields(log.Fields{
			"request_id": cbk.RequestId,
		}).Error("No external ID found in message.")
		// This is sort of my janky way of trying to force a rollback if someone
		// deletes the role in the stack. That's the only circumstance when I
		// force an update failure and a rollback.
		return s.putCallbackResponse(makeResponse(cbk, false), cbk.ResponseURL)
	}

	log.WithFields(log.Fields{
		"request_id":  cbk.RequestId,
		"external_id": externalID,
	}).Info("Handling update request.")
	return s.putCallbackResponse(makeResponse(cbk, true), cbk.ResponseURL)
}

func (s *service) handleCFDeleteRequest(cbk CallbackMessage) error {
	externalID := cbk.ResourceProperties.RoleExternalID
	if externalID == "" {
		log.WithFields(log.Fields{
			"request_id": cbk.RequestId,
		}).Error("No external ID found in message.")
		return s.putCallbackResponse(makeResponse(cbk, false), cbk.ResponseURL)
	}

	account, err := s.db.GetAccountByExternalID(externalID)
	if err != nil {
		return err
	}

	if account == nil {
		log.WithFields(log.Fields{
			"request_id":  cbk.RequestId,
			"external_id": externalID,
		}).WithError(err).Error("No account found.")
		// Go ahead and ACK the delete request.
		return s.putCallbackResponse(makeResponse(cbk, true), cbk.ResponseURL)
	}

	account.Active = false
	err = s.db.UpdateAccount(account)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"request_id":  cbk.RequestId,
		"external_id": externalID,
		"customer_id": account.CustomerID,
	}).Info("Deactivating account on delete request.")

	// firstly, remove any creds we have from the cache
	s.lru.Remove(account.ID)

	stack, err := s.db.GetStack(account.CustomerID, account.ExternalID)
	if err != nil {
		return err
	}

	if stack == nil {
		log.WithFields(log.Fields{
			"request_id":  cbk.RequestId,
			"external_id": externalID,
			"customer_id": account.CustomerID,
		}).Error("Unable to find stack for customer.")
	} else {
		err = s.db.DeleteStack(stack)
		if err != nil {
			return err
		}
	}

	return s.putCallbackResponse(makeResponse(cbk, true), cbk.ResponseURL)
}

func (s *service) handleCFNCallback(msg *CFNotification) error {
	if msg == nil {
		return fmt.Errorf("Recieved nil message")
	}

	if msg.Type != "Notification" {
		return fmt.Errorf("Cannot handle message type: %s", msg.Type)
	}

	cbkMessage := CallbackMessage{}
	err := json.Unmarshal([]byte(msg.Message), &cbkMessage)
	if err != nil {
		return err
	}

	switch cbkMessage.RequestType {
	case "Create":
		return s.handleCFCreateRequest(cbkMessage)
	case "Update":
		return s.handleCFUpdateRequest(cbkMessage)
	case "Delete":
		return s.handleCFDeleteRequest(cbkMessage)
	default:
		return fmt.Errorf("Cannot handle request type: %s", cbkMessage.RequestType)
	}
}
