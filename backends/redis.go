package gilmour

import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"gopkg.in/gilmour-libs/gilmour-go.v0/protocol"
	"strings"
)

const defaultErrorQueue = "gilmour.errorqueue"
const defaultErrorTopic = "gilmour.errors"
const defaultHealthTopic = "gilmour.health"
const defaultResponseTopic = "gilmour.response"
const defaultIdentKey = "gilmour.known_host.health"
const defaultErrorBuffer = 9999

type Redis struct {
	conn   redis.Conn
	pubsub redis.PubSubConn
}

func (self *Redis) getErrorTopic() string {
	return defaultErrorTopic
}

func (self *Redis) getErrorQueue() string {
	return defaultErrorQueue
}

func (self *Redis) reportError(method string, message *protocol.Error) (err error) {
	switch method {
	case protocol.PUBLISH:
		err = self.publish(self.getErrorTopic(), message)

	case protocol.QUEUE:
		msg, merr := message.Marshal()
		if merr != nil {
			err = merr
			return
		}

		queue := self.getErrorQueue()
		self.conn.Send("LPUSH", queue, string(msg))
		self.conn.Send("LTRIM", queue, 0, defaultErrorBuffer)

		_, err = self.conn.Receive()

	}

	return err
}

func (self *Redis) unsubscribe(topic string) (err error) {
	if strings.HasSuffix(topic, "*") {
		err = self.pubsub.PUnsubscribe(topic)
	} else {
		err = self.pubsub.PSubscribe(topic)
	}

	return
}

func (self *Redis) subscribe(topic string) (err error) {
	if strings.HasSuffix(topic, "*") {
		err = self.pubsub.PSubscribe(topic)
	} else {
		err = self.pubsub.Subscribe(topic)
	}

	return
}

func (self *Redis) responseTopic(sender string) string {
	return defaultResponseTopic + "." + sender
}

func (self *Redis) publish(topic string, message interface{}) error {
	switch t := message.(type) {
	case string:
		// Just use t
	case protocol.Messenger:
		msg, err := t.Marshal()
		if err != nil {
			return err
		}
	default:
		return errors.New("Message can only be String or protocol.Messenger")
	}

	return nil
}
