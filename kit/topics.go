package kit

import "fmt"

func HealthTopic(ident string) string {
	return fmt.Sprintf("gilmour.health.%v", ident)
}

func HealthIdent() string {
	return "gilmour.known_host.health"
}

func ResponseTopic(sender string) string {
	return fmt.Sprintf("gilmour.response.%s", sender)
}

func RequestTopic(topic string) string {
	return fmt.Sprintf("gilmour.request.%v", topic)
}

func SlotTopic(topic string) string {
	return fmt.Sprintf("gilmour.slot.%v", topic)
}
