package backends

import (
	"errors"
	"fmt"
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

func MakeRedis(host string) *Redis {
	engine := Redis{}
	engine.conn = GetConn(host)
	_, err := engine.conn.Do("PING")
	if err != nil {
		panic(err)
	}

	engine.pubsub = redis.PubSubConn{Conn: engine.conn}
	return &engine
}

type Redis struct {
	conn   redis.Conn
	pubsub redis.PubSubConn
}

func (self *Redis) getErrorTopic() string {
	return defaultErrorTopic
}

func (self *Redis) AcquireGroupLock(group, sender string) bool {
	key := sender + group

	val, err := self.conn.Do("SET", key, key, "NX", "EX", "600")
	if err != nil {
		return false
	}

	if val == nil {
		return false
	}

	return true
}

func (self *Redis) getErrorQueue() string {
	return defaultErrorQueue
}

func (self *Redis) ReportError(method string, message *protocol.Error) (err error) {
	switch method {
	case protocol.PUBLISH:
		err = self.Publish(self.getErrorTopic(), message)

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

func (self *Redis) Unsubscribe(topic string) (err error) {
	if strings.HasSuffix(topic, "*") {
		err = self.pubsub.PUnsubscribe(topic)
	} else {
		err = self.pubsub.PSubscribe(topic)
	}

	return
}

func (self *Redis) Subscribe(topic string) (err error) {
	if strings.HasSuffix(topic, "*") {
		err = self.pubsub.PSubscribe(topic)
	} else {
		err = self.pubsub.Subscribe(topic)
	}

	return
}

func (self *Redis) ResponseTopic(sender string) string {
	return defaultResponseTopic + "." + sender
}

func (self *Redis) Publish(topic string, message interface{}) (err error) {
	var msg string
	switch t := message.(type) {
	case string:
		msg = t
	case protocol.Messenger:
		msg2, err2 := t.Marshal()
		if err != nil {
			err = err2
		} else {
			msg = string(msg2)
		}
	default:
		err = errors.New("Message can only be String or protocol.Messenger")
	}

	if err != nil {
		_, err = self.conn.Do("PUBLISH", topic, msg)
	}

	return
}

func (self *Redis) RegisterIdent(uuid string) error {
	_, err := self.conn.Do("HSET", defaultIdentKey, uuid, "true")
	return err
}

func (self *Redis) UnregisterIdent(uuid string) error {
	_, err := self.conn.Do("HDEL", defaultIdentKey, uuid)
	return err
}

func (self *Redis) Start() chan *protocol.Message {
	return self.setupListeners()
}

func (self *Redis) Stop() {
}

func (self *Redis) setupListeners() chan *protocol.Message {
	sink := make(chan *protocol.Message) //Add a buffer of 50 messages?

	go func() {
		for {
			switch v := self.pubsub.Receive().(type) {
			case redis.PMessage:
				msg := &protocol.Message{"pmessage", v.Channel, v.Data, v.Pattern}
				sink <- msg
			case redis.Message:
				msg := &protocol.Message{"message", v.Channel, v.Data, v.Channel}
				sink <- msg
			case redis.Subscription:
				fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
			case redis.Pong:
				fmt.Printf("%s: %s %d\n", "Pong", nil, v.Data)
			case error:
				fmt.Printf(v.Error())
			}
		}
	}()

	return sink
}
