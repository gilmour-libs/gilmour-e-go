package protocol

import (
	"encoding/json"
)

type jsonData struct {
	Data   json.RawMessage `json:"data"`
	Code   int             `json:"code"`
	Sender string          `json:"sender"`
}

type GilmourResponse struct {
	jsonData
}

func (self *GilmourResponse) GetData(t interface{}) error {
	return json.Unmarshal(self.jsonData.Data, t)
}

func (self *GilmourResponse) RawData() []byte {
	return self.jsonData.Data
}

func (self *GilmourResponse) GetCode() int {
	return self.jsonData.Code
}

func (self *GilmourResponse) GetSender() string {
	return self.jsonData.Sender
}

func ParseResponse(data string) (resp *GilmourResponse, err error) {
	err = json.Unmarshal([]byte(data), &resp.jsonData)
	return
}
