package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/opsee/spanx/store"
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
	ResourceType       string                     `json:"ResourceType"`
	ResourceProperties CallbackResourceProperties `json:"ResourceProperties"`
}

type CallbackResourceProperties struct {
	ServiceToken   string `json:"ServiceToken"`
	RoleExternalID string `json:"RoleExternalID"`
	RoleARN        string `json:"RoleARN"`
	StackName      string `json:"StackName"`
	StackID        string `json:"StackID"`
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

func (s *service) handleCFCreateRequest(cbk CallbackMessage) error {
	var (
		stack      *store.Stack
		customerID string
		err        error
	)

	externalID := cbk.ResourceProperties.RoleExternalID
	if externalID == "" {
		// Emit failure event to CFN
	}

	// 1. Get the Account object from the DB
	account, err := s.db.GetAccountByExternalID(externalID)
	if err != nil {
		err = errors.New("Error getting external ID")
		goto EMIT_RESPONSE
	}

	// 2. If no account object, return create failed.
	if account == nil {
		err = errors.New("External ID not found.")
		goto EMIT_RESPONSE
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
		if !stack.Active {
			newStack := *stack
			newStack.StackID = cbk.StackId
			newStack.StackName = cbk.ResourceProperties.StackName
			newStack.Active = true
			s.db.UpdateStack(stack, &newStack)
			goto EMIT_RESPONSE
		}
	}

	if cbk.StackId == "" {
		return errors.New("Callback message contained no Stack ID.")
	}

	if cbk.ResourceProperties.StackName == "" {
		return errors.New("Callback contained no stack name.")
	}

	stack = &store.Stack{
		StackID:    cbk.StackId,
		StackName:  cbk.ResourceProperties.StackName,
		CustomerID: customerID,
		ExternalID: externalID,
		Active:     true,
	}

	err = s.db.PutStack(stack)
	if err != nil {
		return err
	}

	if !account.Active {
		newAccount := *account
		newAccount.Active = true
		err = s.db.UpdateAccount(account, &newAccount)
		if err != nil {
			return err
		}
	}

EMIT_RESPONSE:
	response := &CallbackResponse{
		LogicalResourceId: cbk.LogicalResourceId,
		RequestId:         cbk.RequestId,
		StackId:           cbk.StackId,
	}

	if err != nil {
		response.Status = "FAILURE"
		response.Reason = err.Error()
	} else {
		response.Status = "SUCCESS"
		response.PhysicalResourceId = externalID
	}

	return s.putCallbackResponse(response, cbk.ResponseURL)
}

func (s *service) handleCFUpdateRequest(cbk CallbackMessage) error {
	return nil
}

func (s *service) handleCFDeleteRequest(cbk CallbackMessage) error {
	return nil
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
