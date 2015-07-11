package protocol

import (
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

func (self *GilmourError) GetRequestData() string {
	return self.requestData
}

func (self *GilmourError) GetUserData() string {
	return self.userData
}

func (self *GilmourError) GetCode() int {
	return self.code
}

func (self *GilmourError) GetBacktrace() interface{} {
	return self.backtrace
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
