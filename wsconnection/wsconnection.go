package wsconnection

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
	"fmt"
	"encoding/json"
	"strings"
	"regexp"
	"github.com/eluleci/lightning/roothub"
	"github.com/eluleci/lightning/message"
	"github.com/eluleci/lightning/util"
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
}

func CreateConnection(w http.ResponseWriter, r *http.Request) (c Connection) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	c.ws = ws
	return
}

func (c *Connection) Run() {
	defer func() {
		fmt.Println("Connection is teared down. Unsubscribing from channels #", len(c.subscriptions))
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
		fmt.Println("Closing readPump() of connection.")
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
		fmt.Println("Connection received message ", string(m))

		var msg message.Message

		err = json.Unmarshal([]byte(string(m[:])), &msg)
		if err != nil {
			fmt.Println("Error while parsing message: ", err)
			answer := util.CreateErrorMessage(msg.Rid, 400, "Error while parsing message")
			c.send <- answer    // sending the answer back to the connection
		} else {

			if msg.Command == "disconnect" {
				return
			}

			// making sure that the resource has a correct path (ex: "/type/id/type/id")
			msg.Res = "/"+strings.Trim(msg.Res, "/")

			// checking that the resource path is valid path
			matched, err := regexp.MatchString("^(\\/{1}[0-9a-zA-Z]+)+$", msg.Res)
			if !matched || err != nil {
				fmt.Println("Error while validating resource path")
				answer := util.CreateErrorMessage(msg.Rid, 400, "Given resource path is not a valid path.")
				c.send <- answer    // sending the answer back to the connection
			} else {

				rw := message.RequestWrapper{
					msg.Res,
					msg,
					c.send,
					c.subscribed,
				}

				var inbox chan message.RequestWrapper
				subscription, exists := c.subscriptions[msg.Res]
				if exists {
					fmt.Println("Connection has subscription for " + msg.Res)
					inbox = subscription.InboxChannel
				} else {
					for k, v := range c.subscriptions {
						if strings.Index(msg.Res, k) > -1 {
							fmt.Println("Connection has subscription for a parent of " + msg.Res)
							inbox = v.InboxChannel
							break
						}
					}
					if inbox == nil {
						fmt.Println("Connection has no subscription for " + msg.Res)
						inbox = roothub.RootHub.Inbox
					}
				}
				go func() {
					inbox <- rw
				}()
			}
		}
	}
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
			fmt.Println("Closing writePump() of connection.")
			c.tearDown <- true
		}()
		ticker.Stop()
	}()
	for {
		select {
		case message, ok := <-c.send:
			m, _ := json.Marshal(message)

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

