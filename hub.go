// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	_"fmt"
	"fmt"
)

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type Hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Handle the last message in the given connection
	bindNew chan string

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection

	lightning Lightning
}
/*
var h = hub{
	broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
	lightning:   Lightning{
		make([]Profile, 0),
		make(chan ConnectionMessage),
	},
}*/

func (h *Hub) run() {
	go h.lightning.run()

	for {
		select {
		case c := <-h.register:
			fmt.Println("registering a connection to hub")
			h.connections[c] = true

		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}

		case ref := <-h.bindNew:

			// bind all connections to new item
			for c := range h.connections {
				cm := ConnectionBinding{ref, c, Message{}}
				bind <- cm
			}

		case m := <-h.broadcast:

			fmt.Println("broadcasting message: ", string(m))

			// sending the message to all connections
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					close(c.send)
					delete(h.connections, c)
				}
			}
		}
	}
}
