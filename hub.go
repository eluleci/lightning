// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan *connection

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

var h = hub{
	broadcast:   make(chan *connection),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true

			// returning the current state to the new connection
//			c.send <-  GetCurrentState()

		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		case c := <-h.broadcast:

			fmt.Println("@")


			// getting the message from inbox
			m := <- c.inbox

			answer, broadcastMessage := ProcessMessage(m)

			if len(answer) > 1 {
				// generating answer and sending it back to the connection
				c.send <- answer
			}

			if len(broadcastMessage) > 1 {
				// forwarding the broadcast message to all other connections
				for oc := range h.connections {
//					if oc != c {
						select {
						case oc.send <- broadcastMessage:
						default:
							close(oc.send)
							delete(h.connections, oc)
						}
//					}
				}
			}
		}
	}
}
