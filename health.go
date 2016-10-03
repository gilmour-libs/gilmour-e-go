package gilmour

import (
	"strings"

	"gopkg.in/gilmour-libs/gilmour-e-go.v5/kit"
)

func subscribeHealth(g *Gilmour) {
	health_topic := kit.HealthTopic(g.getIdent())
	handlerOpts := NewHandlerOpts().SetGroup("exclusive")

	g.ReplyTo(health_topic, func(r *Request, w *Message) {
		topics := []string{}

		resp_topic := kit.ResponseTopic("")

		for t, _ := range g.getAllSubscribers() {
			if strings.HasPrefix(t, resp_topic) || strings.HasPrefix(t, health_topic) {
				//Do Nothing, these are internal topics
			} else {
				topics = append(topics, t)
			}
		}

		w.SetData(topics)
	}, handlerOpts)
}
