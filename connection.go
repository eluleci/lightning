package main

import (
	"github.com/gorilla/websocket"
	"time"
	"fmt"
	"encoding/json"
	"strings"
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
	send chan Message

	subscribed chan Subscription

	subscriptions  map[string]Subscription

	tearDown chan bool
}

func (c *Connection) run() {
	// tearDownCount is used for waiting both writePump and readPump jobs to be down for closing all channels
	tearDownCount := 2

	defer func() {
		fmt.Println("Connection is teared down.")
		// un-registering from all hubs
		for _, h := range c.subscriptions {
			h.unsubscriptionChannel <- c.send
		}
		c.ws.Close()
		close(c.send)
		close(c.subscribed)
		close(c.tearDown)
	}()

	go c.writePump()
	go c.readPump()

	for {
		select {
		case subscription := <-c.subscribed:
			fmt.Println("Connection subscribed to res: " + subscription.res)
			c.subscriptions[subscription.res] = subscription
		case down := <-c.tearDown:
			if down {
				tearDownCount--
				if tearDownCount == 0 {
					return
				}
			}
		}
	}

}

// readPump pumps messages from the websocket connection to the hub.
func (c *Connection) readPump() {
	defer func() {
		go func() {
			if c.tearDown != nil {
				c.tearDown <- true
			}
		}()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, m, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		fmt.Println("Connection received message")
		fmt.Println(string(m))

		var message Message

		err = json.Unmarshal([]byte(string(m[:])), &message)
		if err != nil {
			fmt.Println("error while parsing message: ", err)
			answer := createErrorMessage(message.Rid, 400, "Error while parsing message")
			c.send <- answer    // sending the answer back to the connection
		} else {
			// making sure that the resource has a correct path (ex: "/type/id/type/id")
			message.Res = "/"+strings.Trim(message.Res, "/")

			rw := RequestWrapper{
				message.Res,
				message,
				c.send,
				c.subscribed,
			}

			subscription, exists := c.subscriptions[message.Res]
			if exists {
				fmt.Println("Connection has subscription for " + message.Res)
				subscription.inboxChannel <- rw
			} else {
				var inbox chan RequestWrapper
				for k, v := range c.subscriptions {
					if strings.Index(message.Res, k) > -1 {
						fmt.Println("Connection has subscription for a parent of " + message.Res)
						inbox = v.inboxChannel
						break
					}
				}
				if inbox == nil {
					fmt.Println("Connection has no subscription for " + message.Res)
					inbox = rootHub.inbox
				}
				inbox <- rw
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
			if c.tearDown != nil {
				c.tearDown <- true
			}
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

