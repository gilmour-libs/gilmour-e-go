package proto

import (
	"encoding/json"
	"time"
)

type GilmourError struct {
	topic        string
	requestData  string
	userData     string
	sender       string
	multiProcess bool
	timestamp    string
	backtrace    interface{}
	code         int
}

func (self *GilmourError) GetTopic() string {
	return self.topic
}

func (self *GilmourError) GetSender() string {
	return self.sender
}

func (self *GilmourError) GetCode() int {
	return self.code
}

func (self *GilmourError) Marshal() ([]byte, error) {
	eStruct := struct {
		Topic        string      `json:"topic"`
		RequestData  string      `json:"request_data"`
		UserData     string      `json:"user_data"`
		Sender       string      `json:"sender"`
		MultiProcess bool        `json:"multi_process"`
		Timestamp    string      `json:"timestamp"`
		Backtrace    interface{} `json:"backtrace"`
		Code         int         `json:"code"`
	}{self.topic, self.requestData, self.userData, self.sender, false,
		self.timestamp, self.backtrace, self.code,
	}

	return json.Marshal(eStruct)
}

func MakeError(
	code int,
	sender, topic, requestData, userData, backtrace string,
) *GilmourError {

	return &GilmourError{
		code:        code,
		topic:       topic,
		requestData: requestData,
		userData:    userData,
		backtrace:   backtrace,
		timestamp:   time.Now().Format(time.RFC3339),
		sender:      sender,
	}
}
