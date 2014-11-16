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

var rootHub Hub
var connections []*Connection

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
	c := &Connection{send: make(chan Message, 256), ws: ws}
	connections = append(connections, c)
	fmt.Println("got new connection. #", len(connections))
	c.run()
}

func main() {

	rootHub = createHub("/")
	go rootHub.run()

	flag.Parse()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

/*
func runMain() {

	for {
		select {
		case cm := <-bind:

			//			fmt.Println("binding connection in main")
			resDomain := cm.message.Res
			res := cm.message.Res
			id := cm.message.Body["id"]
			if id != nil {
				res += "/"+id.(string)
			}
			fmt.Println("searching hub for resource: ", res)

			hub, ok := referenceHubMap[res]

			if !ok {
				// creating a new hub for object
//				hub = createHubForResource(res)

				// checking hub for domain
				hub, ok := referenceHubMap[resDomain]
				if !ok {
					// creating a new hub for domain
					fmt.Println("No hub found for domain: ", resDomain)

					var subscribe chan RequestWrapper
					var appendHub chan RequestWrapper
					go runHub(subscribe, appendHub)
					referenceHubMap[resDomain] = subscribe

					m := Message{}
					m.Res = resDomain
					m.Cmd = "create"
					createDomainRequest := RequestWrapper {
						nil,
						m,
					}
					subscribe <- createDomainRequest        // request for creating empty list
				}

				var subscribe chan RequestWrapper
				var appendHub chan RequestWrapper
				go runHub(subscribe, appendHub)
				referenceHubMap[res] = subscribe
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
}*/
/*

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
			make(chan RequestWrapper),
			broadcast,
			bindNew,
		},
	}
	go hub.run()
	return hub
}
*/
