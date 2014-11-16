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
	pongWait = 60 * time.Second

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
}

func (c *Connection) run() {
	go c.writePump()
	c.readPump()
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Connection) readPump() {
	defer func() {
		// un-registering from all hubs
		//		for _, h := range c.referencedHubMap {
		//			h.unregister <- c
		//		}
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, m, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		fmt.Println("connection received message")
		fmt.Println(string(m))

		var message Message

		err = json.Unmarshal([]byte(string(m[:])), &message)
		if err != nil {
			fmt.Println("error while parsing message: ", err)
			answer := Message{}
			answer.Rid = message.Rid
			answer.Status = 400
			c.send <- answer    // sending the answer back to the connection
			return
		}
		// making sure that the resource has a correct path (ex: "/type/id/type/id")
		message.Res = "/"+strings.Trim(message.Res, "/")

		rw := RequestWrapper{
			message.Res,
			message,
			c.send,
		}
		rootHub.inbox <- rw

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
		ticker.Stop()
		c.ws.Close()
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

