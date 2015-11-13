package protocol

import (
	"encoding/json"
)

type RecvRequestd struct {
	Data   json.RawMessage `json:"data"`
	Code   int             `json:"code"`
	Sender string          `json:"sender"`
}

func (self *RecvRequestd) GetData(t interface{}) error {
	return json.Unmarshal(self.Data, t)
}

func (self *RecvRequestd) RawData() []byte {
	return self.Data
}

func (self *RecvRequestd) GetCode() int {
	return self.Code
}

func (self *RecvRequestd) GetSender() string {
	return self.Sender
}

func ParseResponse(data interface{}) (resp *RecvRequestd, err error) {
	var msg []byte

	switch t := data.(type) {
	case string:
		msg = []byte(t)
	case []byte:
		msg = t
	case json.RawMessage:
		msg = t
	}

	err = json.Unmarshal(msg, &resp)
	return
}
