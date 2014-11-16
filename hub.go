// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strings"
)

// this struct is used for connections to send a message to a hub that they are not subscribed before
// this is created from a connection and sent to rootHub's inbox channel. rootHub finds the related hub and that hub
// executes the message inside the request. also, the connection receives a Subscription object from that hub
type RequestWrapper struct {
	res      string
	message  Message
	listener chan Message
}

type Hub struct {
	res          string
	model        ModelHolder
	children    map[string]Hub
	subscribers map[chan Message]bool
	inbox        chan RequestWrapper
	broadcast    chan Message
	subscribe    chan chan Message
	unsubscribe  chan chan Message
}

// this strut is used for notifying that a connection is subscribed to a hub. this can happen in two ways:
// 1: if connection sends a request wrapper to rootHub, rootHub finds the requested hub and gives the request. then the
// message is executed by the related hub and also the connection is subscribed to that hub
// 2: if a connection is subscribed to a list of resource, the connection is automatically subscribed to that list if a
// new item is created inside that list
type Subscription struct {
	res                   string
	inboxChannel          chan Message
	broadcastChannel      chan Message
	unsubscriptionChannel chan Message
}

func (h *Hub) run() {

	fmt.Println(h.res + ":  Started running.")

	for {
		select {
		case requestWrapper := <-h.inbox:
			fmt.Println(h.res+": received message: ", requestWrapper.message)

			if requestWrapper.res == h.res {
				// if the resource of that message is this hub's resource
				fmt.Println(h.res + ": responsible of message.")
				fmt.Println(h.model.model)

				if h.model.model["id"] != nil {
					// regular message, forward to model
					fmt.Println(h.res + ": Forwarding message to model handler")
					h.model.handle <- requestWrapper

				} else if (requestWrapper.message.Command == "post" && requestWrapper.message.Body["id"] != nil) {
					// this is object initialisation message
					h.model = createModelHolder(h.broadcast)
					go h.model.run()
					h.model.handle <- requestWrapper

				} else if requestWrapper.message.Command == "post" && requestWrapper.message.Body["id"] == nil {
					// if command is post and there is no id inside the message, it means it is a message for object
					// creation under this domain
					generatedId := randSeq(32)
					generatedRes := requestWrapper.res + "/" + generatedId

					hub := createHub(generatedRes)
					go hub.run()
					h.children[hub.res] = hub
					fmt.Println(h.res+": Created a new object for res: ", hub.res)

					requestWrapper.res = generatedRes
					requestWrapper.message.Res = generatedRes
					requestWrapper.message.Body["id"] = generatedId
					hub.inbox <- requestWrapper

				} else if requestWrapper.message.Command == "get" && len(h.children) > 0 {
					// this is a get message of the list
					fmt.Println(h.res + ": Getting list of the models in the resource " + h.res)
					list := make([]map[string]interface{}, len(h.children))
					callback := make(chan Message)

					// sending get messages
					for k, v := range h.children {
						var getMessage Message
						getMessage.Command = "get"
						var rw RequestWrapper
						rw.res = k
						rw.message = getMessage
						rw.listener = callback
						v.inbox <- rw
					}
					// receiving responses (receiving response is done after sending all messages for preventing being
					// blocked by a model)
					for i := 0; i < len(h.children); i++ {
						response := <-callback
						fmt.Println(response.Body)
						list[i] = response.Body
					}
					var answer Message
					answer.Rid = requestWrapper.message.Rid
					answer.Status = 200
					answer.Body = make(map[string]interface{})
					answer.Body["list"] = list
					requestWrapper.listener <- answer
				}
			} else {
				// if the resource belongs to a children hub
				_, childRes := getChildRes(requestWrapper.res, h.res)
				fmt.Println("direct child of " + h.res + " is " + childRes)

				hub, exists := h.children[childRes]
				if !exists {
					//   if children doesn't exists -> create children hub for the resource
					fmt.Println(h.res+": Hub doesn't exists for res: ", requestWrapper.res)
					hub = createHub(childRes)
					go hub.run()
					h.children[childRes] = hub
					fmt.Println(h.res+": Created a hub for res: ", hub.res)
				}
				//   forward message to the children hub
				hub.inbox <- requestWrapper
			}

		case listener := <-h.subscribe:
			fmt.Println(h.res + ": Adding new listener to subscribers of " + h.res)
			h.subscribers[listener] = true

		case listener := <-h.unsubscribe:
			fmt.Println(h.res + ": Removing a listener from subscribers of " + h.res)
			if _, ok := h.subscribers[listener]; ok {
				delete(h.subscribers, listener)
			}
		}
	}

}

func createHub(res string) (h Hub) {
	h.res = res
	h.children = make(map[string]Hub)
	h.subscribers = make(map[chan Message]bool)
	h.inbox = make(chan RequestWrapper)
	h.broadcast = make(chan Message)
	h.subscribe = make(chan chan Message)
	h.unsubscribe = make(chan chan Message)
	return
}

func createModelHolder(broadcastChannel chan Message) (mh ModelHolder) {
	mh.handle = make(chan RequestWrapper)
	mh.broadcastChannel = broadcastChannel
	return
}

func getChildRes(res, parentRes string) (relativePath, fullPath string) {
	res = strings.Trim(res, "/")
	parentRes = strings.Trim(parentRes, "/")
	currentResSize := len(parentRes)
	resSuffix := res[currentResSize:]
	trimmedSuffix := strings.Trim(resSuffix, "/")
	directChild := strings.Split(trimmedSuffix, "/")
	relativePath = directChild[0]
	if len(parentRes) > 0 {
		fullPath = "/"+parentRes+"/"+relativePath
	} else {
		fullPath = "/"+relativePath
	}
	return
}
