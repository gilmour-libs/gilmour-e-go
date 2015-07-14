package protocol

import (
	"encoding/json"
)

type jsonData struct {
	Data   json.RawMessage `json:"data"`
	Code   int             `json:"code"`
	Sender string          `json:"sender"`
}

type RecvRequest struct {
	jsonData jsonData
}

func (self *RecvRequest) GetData(t interface{}) error {
	return json.Unmarshal(self.jsonData.Data, t)
}

func (self *RecvRequest) RawData() []byte {
	return self.jsonData.Data
}

func (self *RecvRequest) GetCode() int {
	return self.jsonData.Code
}

func (self *RecvRequest) GetSender() string {
	return self.jsonData.Sender
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

	err = json.Unmarshal(msg, &resp.jsonData)
	return
}
