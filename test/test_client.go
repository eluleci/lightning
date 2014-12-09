package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"time"
)

var host = "localhost"

//var host = "uekk53169e3d.eluleci.koding.io"
var origin = "http://" + host + ":8080"
var url = "ws://" + host + ":8080/ws"

type Message struct {

	Rid                      int `json:"rid,omitempty"`
	Status                   int `json:"status,omitempty"`
	Res                      string `json:"res,omitempty"`
	Command                  string `json:"cmd,omitempty"`
	Parameters               string `json:"params,omitempty"`
	Body                map[string]interface{} `json:"body,omitempty"`
}

var ws *websocket.Conn

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Defer in main", r)
			startProcess()
		}
	}()

	startProcess()

}

func getTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func startProcess() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered an error. Restarting...", r)
			startProcess()
		}
	}()

	ws, _ = websocket.Dial(url, "", origin)

	sendSetHeadersMessage(ws)

	go readMessages(ws)
	startSendingCreateMessage(ws)
}

func readMessages(ws *websocket.Conn) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in readMessages()")
			readMessages(ws)
		}
	}()

	for {
		var msg = make([]byte, 512)
		size, err := ws.Read(msg)
		if err != nil {
			fmt.Println("EOF error: %s\n", msg)

		} else {
			fmt.Println("Received: ", string(msg[:512]))

			var message Message
			err = json.Unmarshal(msg[:size], &message)
			if err != nil {
				fmt.Println("Error while parsing message: ", err)
				fmt.Println("Error while parsing message: ", msg[:size])
			} else {
				if (message.Status == 200 && message.Body != nil) {
					go startSendingUpdateMessage(ws, message.Res)
				}
			}
		}
	}
}

func sendSetHeadersMessage(ws *websocket.Conn) {
	setHeadersMessage := Message{}
	setHeadersMessage.Rid = 1
	setHeadersMessage.Command = "::setHeaders"
	body := make(map[string]interface{})
	body["X-Parse-Application-Id"] = "oxjtnRnmGUKyM9SFd1szSKzO9wKHGlgC6WgyRpq8"
	body["X-Parse-REST-API-Key"] = "qJcOosDh8pywHdWKkVryWPoQFT1JMyoZYjMvnUul"
	setHeadersMessage.Body = body
	setHeadersMessageByte, jErr := json.Marshal(setHeadersMessage)
	if jErr == nil {
		ws.Write(setHeadersMessageByte)
	}
}

func startSendingCreateMessage(ws *websocket.Conn) {

	createMessage := Message{}
	createMessage.Rid = 1
	createMessage.Res = "/Comment"
	createMessage.Command = "post"
	body := make(map[string]interface{})
	body["content"] = "comment"
	body["likes"] = 0
	createMessage.Body = body

	sendMessagePeriodically(ws, createMessage, 2000)
}

func startSendingUpdateMessage(ws *websocket.Conn, res string) {

	if (len(res) == 0) {
		return
	}
	fmt.Println("Starting sending update messages for ", res)

	updateMessage := Message{}
	updateMessage.Rid = 1
	updateMessage.Res = res
	updateMessage.Command = "put"
	body := make(map[string]interface{})
	body["likes"] = getTimestamp()
	updateMessage.Body = body

	sendMessagePeriodically(ws, updateMessage, 1000)
}

func sendMessagePeriodically(ws *websocket.Conn, message Message, period int64) {

	for {
		//		fmt.Printf("Waiting 100 millisecs")
		time.Sleep(time.Duration(period) * time.Millisecond)

		m, jErr := json.Marshal(message)
		if jErr == nil {
			_, err := ws.Write(m)
			if err != nil {
				fmt.Printf("Broken pipe error.")
				panic("BP")
			}
			//			fmt.Printf("Sent: %s\n", message)
		}
	}
}
