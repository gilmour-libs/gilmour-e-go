package gilmour

import (
	"strings"
)

func subscribeHealth(self *Gilmour) {
	ident := self.GetIdent()
	health_topic := self.backend.HealthTopic(ident)

	self.Subscribe(health_topic, func(r *Request, w *Response) {
		topics := []string{}

		resp_topic := self.backend.ResponseTopic("")

		for t, _ := range self.subscribers {
			if strings.HasPrefix(t, resp_topic) ||
				strings.HasPrefix(t, health_topic) {
				//Do Nothing, these are internal topics
			} else {
				topics = append(topics, t)
			}
		}

		w.Respond(topics)
	}, MakeHandlerOpts().SetGroup("exclusive"))
}
