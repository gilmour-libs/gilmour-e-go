package protocol

import (
	"encoding/json"
	"time"
)

type Error struct {
	topic        string
	requestData  string
	userData     string
	sender       string
	multiProcess bool
	timestamp    string
	backtrace    interface{}
	code         int
}

func (self *Error) GetTopic() string {
	return self.topic
}

func (self *Error) GetSender() string {
	return self.sender
}

func (self *Error) GetRequestData() string {
	return self.requestData
}

func (self *Error) GetUserData() string {
	return self.userData
}

func (self *Error) GetCode() int {
	return self.code
}

func (self *Error) GetBacktrace() interface{} {
	return self.backtrace
}

func (self *Error) Marshal() ([]byte, error) {
	return json.Marshal(self)
}

func MakeError(
	code int,
	sender, topic, requestData, userData, backtrace string,
) *Error {

	return &Error{
		code:        code,
		topic:       topic,
		requestData: requestData,
		userData:    userData,
		backtrace:   backtrace,
		timestamp:   time.Now().Format(time.RFC3339),
		sender:      sender,
	}
}
