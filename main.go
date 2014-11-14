// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"
	"text/template"
	"fmt"
)

var addr = flag.String("addr", ":8080", "http service address")
var homeTempl = template.Must(template.ParseFiles("home.html"))

var bind chan ConnectionRequest
var connections []*connection
var referenceHubMap map[string]Hub

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl.Execute(w, r.Host)
}

// serverWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("err")
		log.Println(err)
		return
	}
	c := &connection{send: make(chan []byte, 256), registered: make(chan registration, 256), ws: ws, referencedHubMap : make(map[string]Hub)}
	connections = append(connections, c)
	fmt.Println("got new connection. #", len(connections))
	//	go c.run()
	go c.writePump()
	go c.readPump()
}

func main() {

	referenceHubMap = make(map[string]Hub)
	bind = make(chan ConnectionRequest, 256)
	go runMain()

	flag.Parse()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func runMain() {

	for {
		select {
		case cm := <-bind:

			//			fmt.Println("binding connection in main")
			res := cm.message.Res
			id := cm.message.Body["id"]
			if id != nil {
				res += "/"+id.(string)
			}
			fmt.Println("searching hub for resource: ", res)

			hub, ok := referenceHubMap[res]

			if !ok {
				// creating a new hub
				hub = createHubForResource(res)

				if id == nil {
					fmt.Println("no hub found for resource DOMAIN ", res)

					m := Message{}
					m.Res = res
					createDomainRequest := ConnectionRequest {
						nil,
						m,
					}
					hub.lightning.handle <- createDomainRequest        // request for creating empty list

				}
				referenceHubMap[res] = hub
				fmt.Println("added new hub to central map")
				for i := range referenceHubMap {
					fmt.Println(i)
				}
			}
		hub.register <- cm.connection
			registration := registration {
				res,
				hub,
			}
			if cm.connection != nil {
				cm.connection.registered <- registration
			}
		hub.lightning.handle <- cm
		}
	}
}

func createHubForResource(res string) Hub {
	fmt.Println("creating hub for resource ", res)
	broadcast := make(chan []byte)
	bindNew := make(chan Message)
	hub := Hub {
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		broadcast:   broadcast,
		bindNew: bindNew,
		connections: make(map[*connection]bool),
		lightning:   Lightning{
			make(map[string]interface{}),
			make(chan ConnectionRequest),
			broadcast,
			bindNew,
		},
	}
	go hub.run()
	return hub
}
