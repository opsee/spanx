package service

import (
	"encoding/json"
	"fmt"
)

const (
	CreateSuccess = "SUCCESS"
	CreateFailed  = "FAILED"
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
	RequestType        string                       `json:"RequestType"`
	ServiceToken       string                       `json:"ServiceToken"`
	ResponseURL        string                       `json:"ResponseURL"`
	StackId            string                       `json:"StackId"`
	RequestId          string                       `json:"RequestId"`
	LogicalResourceId  string                       `json:"LogicalResourceId"`
	ResourceType       string                       `json:"ResourceType"`
	ResourceProperties CFCallbackResourceProperties `json:"ResourceProperties"`
}

type CallbackResourceProperties struct {
	ServiceToken   string `json:"ServiceToken"`
	RoleExternalID string `json:"RoleExternalID"`
	RoleARN        string `json:"RoleARN"`
	StackName      string `json:"StackName"`
	StackID        string `json:"StackID"`
}

type CallbackResponse struct {
	Status            string `json:"Status"`
	LogicalResourceId string `json:"LogicalResourceId"`
	RequestId         string `json:"RequestId"`
	StackId           string `json:"StackId"`
}

type CallbackCreateSuccess struct {
	CallbackResponse
	PhysicalResourceId string            `json:"PhysicalResourceId"`
	Data               map[string]string `json:"Data"`
}

type CallbackCreateFailed struct {
	CallbackResponse
	Reason string `json:"Reason"`
}

type CallbackDeleteSuccess struct {
	PhysicalResourceId string `json:"PhysicalResourceId"`
}

type CallbackDeleteFailed struct {
	CallbackResponse
	Reason             string `json:"Reason"`
	PhysicalResourceId string `json:"PhysicalResourceId"`
}

type CallbackUpdateSuccess struct {
	CallbackResponse
	PhysicalResourceId string            `json:"PhysicalResourceId"`
	Data               map[string]string `json:"Data"`
}

type CallbackUpdateFailed struct {
	CallbackResponse
	Reason             string `json:"Reason"`
	PhysicalResourceId string `json:"PhysicalResourceId"`
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

	resp, err := callbackResponseClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Received bad response code from S3: %i", resp.StatusCode)
	}

	return nil
}

func (s *service) handleCFCreateRequest(cbk CallbackMessage) error {

}

func (s *service) handleCFUpdateRequest(cbk CallbackMessage) error {
}

func (s *service) handleCFDeleteRequest(cbk CallbackMessage) error {
}

func (s *service) handleCFNCallback(msg *CFNotification) error {
	if msg == nil {
		return fmt.Errorf("Recieved nil message")
	}

	if msg.Type != "Notification" {
		return fmt.Errorf("Cannot handle message type: %s", msg.Type)
	}

	cbkMessage := CallbackMessage{}
	err := json.Unmarshal(msg.Message, cbkMessage)
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
