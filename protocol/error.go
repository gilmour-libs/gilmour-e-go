package protocol

import (
	"time"
)

type GilmourError struct {
	Topic        string
	RequestData  string
	UserData     string
	Sender       string
	MultiProcess bool
	Timestamp    string
	Backtrace    interface{}
	Code         int
}

func (self *GilmourError) GetTopic() string {
	return self.Topic
}

func (self *GilmourError) GetSender() string {
	return self.Sender
}

func (self *GilmourError) GetRequestData() string {
	return self.RequestData
}

func (self *GilmourError) GetUserData() string {
	return self.UserData
}

func (self *GilmourError) GetCode() int {
	return self.Code
}

func (self *GilmourError) GetBacktrace() interface{} {
	return self.Backtrace
}

func MakeError(
	code int,
	sender, topic, requestData, userData, backtrace string,
) *GilmourError {

	return &GilmourError{
		Code:        code,
		Topic:       topic,
		RequestData: requestData,
		UserData:    userData,
		Backtrace:   backtrace,
		Timestamp:   time.Now().Format(time.RFC3339),
		Sender:      sender,
	}
}
