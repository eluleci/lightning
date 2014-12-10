package wsconnection

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
	"encoding/json"
	"strings"
	"regexp"
	"github.com/eluleci/lightning/roothub"
	"github.com/eluleci/lightning/message"
	"github.com/eluleci/lightning/util"
	"strconv"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// connection is an middleman between the websocket connection and the hub.
type Connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan message.Message

	subscribed chan message.Subscription

	subscriptions  map[string]message.Subscription

	tearDown chan bool

	headers map[string][]string
}

func CreateConnection(w http.ResponseWriter, r *http.Request) (c Connection) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		util.Log("error", "WSConnection: Error while upgrading to ws connection.")
		return
	}
	c.ws = ws
	return
}

func (c *Connection) Run() {
	defer func() {
		util.Log("debug", "WSConnection: Connection is teared down. Unsubscribing from channels #"+strconv.Itoa(len(c.subscriptions)))
		// un-registering from all hubs
		for _, subscription := range c.subscriptions {
			//			fmt.Println("Unsubscribing from ", subscription.res)
			rw := new(message.RequestWrapper)
			rw.Listener = c.send
			subscription.UnsubscriptionChannel <- *rw
		}
		c.ws.Close()
		close(c.send)
		close(c.subscribed)
		close(c.tearDown)
	}()

	c.send = make(chan message.Message, 256)
	c.subscribed = make(chan message.Subscription, 256)
	c.subscriptions = make(map[string]message.Subscription)
	c.tearDown = make(chan bool)
	c.headers = make(map[string][]string)

	go c.writePump()    // running message wait and send function
	go c.readPump()     // running message receive function

	for {
		select {
		case subscription := <-c.subscribed:
			//			fmt.Println("Connection subscribed to res: " + subscription.res)
			c.subscriptions[subscription.Res] = subscription
		case down := <-c.tearDown:
			// finishes the run() of connection
			if down {
				return
			}
		}
	}

}

func (c *Connection) readPump() {
	defer func() {
		defer func() {
			if r := recover(); r != nil {
				//				fmt.Println("Recovered in connection. Teardown already closed", r)
			}
		}()
		util.Log("debug", "WSConnection: Closing readPump().")
		c.tearDown <- true
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, m, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		util.Log("debug", "WSConnection: Received message "+string(m))

		var msg message.Message
		err = json.Unmarshal([]byte(string(m[:])), &msg)
		if err != nil {
			util.Log("error", "WSConnection: Error while parsing message.")
			c.send <- createErrorMessage(msg.Rid, 400, "Error while parsing message")
		}

		// checking the command of the message
		if len(msg.Command) == 0 || !isCommandValid(msg.Command) {
			util.Log("error", "WSConnection: Error while validating command.")
			c.send <- createErrorMessage(msg.Rid, 400, "Given command is not a valid command.")
			continue
		}

		// checking custom messages first
		// TODO: subscribe to hub
		// TODO: un-subscribe from hub
		// TODO: disconnect
		if msg.Command == "::setHeaders" {

			if success := c.setHeaders(msg.Body); success {
				c.send <- createSuccessMessage(msg.Rid)
			} else {
				c.send <- createErrorMessage(msg.Rid, 400, "Headers couldn't be set.")
			}
			continue
		}

		// checking the res of the message
		isResValid := isResValid(&msg.Res)
		if !isResValid {
			util.Log("error", "WSConnection: Error while validating resource path.")
			c.send <- createErrorMessage(msg.Rid, 400, "Given resource path is not a valid path.")
			continue
		}

		// add the headers that the connection keeps to the current message
		c.appendHeadersToMessage(&msg)

		// generating request wrapper for the message
		requestWrapper := c.constructRequestFromMessage(msg)

		// finding the best hub to send message
		inbox := c.findAppropriateInbox(msg.Res)

		go func() {
			inbox <- requestWrapper
		}()
	}
}

func (c *Connection) setHeaders(headers map[string]interface{}) bool {
	if headers != nil && len(headers) > 0 {
		for headerName, headerValues := range headers {

			// if values array is empty, it is for removing the header
			if len(headerValues.([]interface{})) == 0 {
				delete(c.headers, headerName)
				continue
			}

			// adding values to headerValues array
			for _, value := range headerValues.([]interface{}) {
				c.headers[headerName] = append(c.headers[headerName], value.(string))
			}
		}
		return true
	} else {
		return false
	}
}

func (c *Connection) appendHeadersToMessage(msg *message.Message) {

	if c.headers != nil {
		if msg.Headers == nil {
			msg.Headers = c.headers
			return
		}
		for existingHeaderName, existingValues := range c.headers {
			for messageHeaderName, messageHeaderValues := range msg.Headers {
				if existingHeaderName == messageHeaderName {
					existingValues = append(existingValues, messageHeaderValues...)
				}
			}
		}
	}
}

func (c *Connection) constructRequestFromMessage(msg message.Message) (rw message.RequestWrapper) {
	rw.Res = msg.Res
	rw.Message = msg
	rw.Listener = c.send
	rw.Subscribe = c.subscribed
	return
}

func (c *Connection) findAppropriateInbox(res string) (inbox chan message.RequestWrapper) {
	subscription, exists := c.subscriptions[res]
	if exists {
		util.Log("debug", "WSConnection: Has subscription for "+res)
		inbox = subscription.InboxChannel
	} else {
		for k, v := range c.subscriptions {
			if strings.Index(res, k) > -1 {
				util.Log("debug", "WSConnection: Has subscription for a parent of "+res)
				inbox = v.InboxChannel
				break
			}
		}
		if inbox == nil {
			util.Log("debug", "WSConnection: Has no subscription for "+res)
			inbox = roothub.RootHub.Inbox
		}
	}
	return
}

// write writes a message with the given message type and payload.
func (c *Connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					//					fmt.Println("Recovered in connection. Teardown already closed", r)
				}
			}()
			util.Log("debug", "WSConnection: Closing writePump().")
			c.tearDown <- true
		}()
		ticker.Stop()
	}()
	for {
		select {
		case message, ok := <-c.send:
			m, _ := json.Marshal(message)
			util.Log("debug", "WSConnection: Sending message "+string(m))

			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, m); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func isCommandValid(cmd string) bool {
	return cmd == "::subscribe" || cmd == "::unsubscribe" || cmd == "::setHeaders" || cmd == "::disconnect" || cmd == "get" || cmd == "put" || cmd == "post" || cmd == "delete"
}

func isResValid(res *string) bool {
	if len(*res) == 0 {
		return false
	}
	// making sure that the resource has a correct path (ex: "/type/id/type/id")
	*res = "/"+strings.Trim(*res, "/")

	// checking that the resource path is valid path
	matched, err := regexp.MatchString("^(\\/{1}[0-9a-zA-Z]+)+$", *res)
	return matched && err == nil
}

func createSuccessMessage(rid int) (m message.Message) {
	m.Rid = rid
	m.Status = 200
	return
}

func createErrorMessage(rid, status int, messageContent string) (m message.Message) {
	_ = messageContent
	m.Rid = rid
	m.Status = status
	return
}
