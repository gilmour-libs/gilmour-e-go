package protocol

import (
	"encoding/json"
)

type RecvRequest struct {
	Data   json.RawMessage `json:"data"`
	Code   int             `json:"code"`
	Sender string          `json:"sender"`
}

func (self *RecvRequest) GetData(t interface{}) error {
	return json.Unmarshal(self.Data, t)
}

func (self *RecvRequest) RawData() []byte {
	return self.Data
}

func (self *RecvRequest) GetCode() int {
	return self.Code
}

func (self *RecvRequest) GetSender() string {
	return self.Sender
}

func ParseResponse(data interface{}) (resp *RecvRequest, err error) {
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
