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
	c := &Connection{ ws: ws, send: make(chan Message, 256), subscribed: make(chan Subscription, 256), subscriptions: make(map[string]Subscription), tearDown: make(chan bool)}
	fmt.Println("Got new connection.")
	c.run()
}

func main() {

	rootHub = createHub("/", nil)
	go rootHub.run()

	flag.Parse()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
