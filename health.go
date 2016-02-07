package gilmour

import (
	"fmt"
	"strings"
)

func healthTopic(ident string) string {
	return fmt.Sprintf("gilmour.health.%v", ident)
}

func subscribeHealth(g *Gilmour) {
	health_topic := healthTopic(g.getIdent())
	handlerOpts := NewHandlerOpts().SetGroup("exclusive")

	g.ReplyTo(health_topic, func(r *Request, w *Message) {
		topics := []string{}

		resp_topic := responseTopic("")

		for t, _ := range g.getAllSubscribers() {
			if strings.HasPrefix(t, resp_topic) || strings.HasPrefix(t, health_topic) {
				//Do Nothing, these are internal topics
			} else {
				topics = append(topics, t)
			}
		}

		w.Send(topics)
	}, handlerOpts)
}
