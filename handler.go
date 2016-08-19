package gilmour

/*
Handler is where your core logic resides, whether it is a Request-Reply or a
Signal-Slot pattern.

For sake of consistency and inerchangability handlers of both Request-Reply &
Signal-Slot must be of the underlying type which accept two arguments;
Incoming *Request and outgoing **Message. (You can read more on Request & *Message in their individual documentations)

Please note that in case of Signal-Slot, as the design pattern suggests,
outgoing **Message is redundant nd any data written to *Message will not be
transmitted to any reciver, because there is no reciver awaiting the response.
*/
type handler func(*Request, *Message)
type RequestHandler func(*Request, *Message)
type SlotHandler func(*Request)
