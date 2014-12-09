package main

import (
	"flag"
	"log"
	"net/http"
	"text/template"
	"github.com/eluleci/lightning/roothub"
	"github.com/eluleci/lightning/wsconnection"
	"github.com/eluleci/lightning/node"
	"github.com/eluleci/lightning/util"
)

var addr = flag.String("addr", ":8080", "http service address")
var homeTempl = template.Must(template.ParseFiles("home.html"))

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

	c := wsconnection.CreateConnection(w, r)
	util.Log("debug", "Main: Got new connection.")
	c.Run()
}

func main() {

	roothub.RootHub = node.CreateHub("/", nil, nil)
	go roothub.RootHub.Run()

	flag.Parse()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
