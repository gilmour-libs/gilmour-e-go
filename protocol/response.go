package protocol

import (
	"encoding/json"
)

type jsonData struct {
	Data   json.RawMessage `json:"data"`
	Code   int             `json:"code"`
	Sender string          `json:"sender"`
}

type RequestResponse struct {
	jsonData jsonData
}

func (self *RequestResponse) GetData(t interface{}) error {
	return json.Unmarshal(self.jsonData.Data, t)
}

func (self *RequestResponse) RawData() []byte {
	return self.jsonData.Data
}

func (self *RequestResponse) GetCode() int {
	return self.jsonData.Code
}

func (self *RequestResponse) GetSender() string {
	return self.jsonData.Sender
}

func ParseResponse(data string) (resp *RequestResponse, err error) {
	err = json.Unmarshal([]byte(data), &resp.jsonData)
	return
}
