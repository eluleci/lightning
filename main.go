package main

import (
	"flag"
	"log"
	"io"
	"net/http"
	"text/template"
	"github.com/eluleci/lightning/roothub"
	"github.com/eluleci/lightning/wsconnection"
	"github.com/eluleci/lightning/node"
	"github.com/eluleci/lightning/util"
	"github.com/eluleci/lightning/message"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
)

var addr = flag.String("addr", ":8080", "http service address")
var homeTempl = template.Must(template.ParseFiles("home.html"))

func serveHome(w http.ResponseWriter, r *http.Request) {
	fmt.Println("serveHome")

	if r.URL.Path != "/panel" {
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

func serveApi(w http.ResponseWriter, r *http.Request) {
	fmt.Println("serveApi")

	vars := mux.Vars(r)
	res := vars["res"]

	var requestWrapper message.RequestWrapper
	requestWrapper.Res = "/"+res
	requestWrapper.Message.Res = "/"+res
	requestWrapper.Message.Command = r.Method
	requestWrapper.Message.Headers = r.Header
	rBody, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(rBody, &requestWrapper.Message.Body)

	bytes, err := json.Marshal(requestWrapper.Message)
	if err != nil {
		http.Error(w, "Internal server error", 500)
	}

	responseChannel := make(chan message.Message)
	requestWrapper.Listener = responseChannel

	roothub.RootHub.Inbox <- requestWrapper

	response := <- responseChannel
	fmt.Println(response)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	io.WriteString(w, string(bytes))


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

	r := mux.NewRouter()
	r.HandleFunc("/api/{res}", serveApi)
	r.HandleFunc("/panel", serveHome)
	r.HandleFunc("/ws", serveWs)
	http.Handle("/", r)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

