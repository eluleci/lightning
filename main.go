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
	"github.com/gorilla/mux"
	"io/ioutil"
	"time"
)

var addr = flag.String("addr", ":8080", "http service address")
var parsePanel = template.Must(template.ParseFiles("parsePanel.html"))
var maidanPanel = template.Must(template.ParseFiles("maidanPanel.html"))

func serveParsePanel(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/parsePanel" {
		http.Error(w, "Not found.", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed.", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	parsePanel.Execute(w, r.Host)
}

func serveMaidanPanel(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/maidanPanel" {
		http.Error(w, "Not found.", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed.", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	maidanPanel.Execute(w, r.Host)
}

func serveHTTP(w http.ResponseWriter, r *http.Request) {

	start := time.Now()
	util.Log("debug", "HTTP: Received request")

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
	util.Log("info", "HTTP: Received request: "+r.Method)

	responseChannel := make(chan message.Message)
	requestWrapper.Listener = responseChannel

	roothub.RootHub.Inbox <- requestWrapper

	response := <-responseChannel
	response.Status = 0    // there is no need to status in http response
	bytes, err = json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal server error", 500)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	io.WriteString(w, string(bytes))

	elapsed := time.Since(start)
	util.Log("info", "HTTP: Response sent in "+elapsed.String())

	close(responseChannel)
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
	r.HandleFunc("/parsePanel", serveParsePanel)
	r.HandleFunc("/maidanPanel", serveMaidanPanel)
	r.HandleFunc("/http/{res:[a-zA-Z0-9/]+}", serveHTTP)
	r.HandleFunc("/ws", serveWs)
	http.Handle("/", r)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

