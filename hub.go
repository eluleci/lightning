package main

import (
	"fmt"
	"strings"
)

type Hub struct {
	res          string
	model        ModelHolder
	children     map[string]Hub
	subscribers  map[chan Message]bool
	inbox        chan RequestWrapper
	broadcast    chan RequestWrapper
	subscribe    chan chan Message
	unsubscribe  chan chan Message
}

func (h *Hub) run() {

	fmt.Println(h.res + ":  Started running.")

	for {
		select {
		case requestWrapper := <-h.inbox:
			fmt.Println(h.res+": Received message: ", requestWrapper.message)

			if requestWrapper.res == h.res {
				// if the resource of that message is this hub's resource
				fmt.Println(h.res + ": responsible of message.")

				if h.model.model["id"] != nil {
					// model exists, so forward message to model handler
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

					requestWrapper.message.Body["id"] = generatedId

					// broadcasting the object creation. we need to create a new request wrapper because we are changing
					// the res of the request wrapper when sending it to newly created hub. but subscribers of this
					// domain will expect a broadcast message with the resource path of this domain
					requestWrapperForBroadcast := requestWrapper
					go func() {
						h.broadcast <- requestWrapperForBroadcast
					}()

					// sending request to the new hub with new res
					requestWrapper.message.Res = generatedRes
					requestWrapper.res = generatedRes
					hub.inbox <- requestWrapper
					continue

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
					// blocked by a get message)
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
				} else {
					requestWrapper.listener <- createErrorMessage(requestWrapper.message.Rid, 404, "Not found.")
					continue
				}

				// subscribing the request sender if there is a subscription channel inside the request
				if requestWrapper.subscribe != nil {
					_, exists := h.subscribers[requestWrapper.listener]
					if !exists {
						fmt.Println("Adding connection to subscription list.")
						subscription := Subscription {
							h.res,
							h.inbox,
							h.broadcast,
							h.unsubscribe,
						}
						requestWrapper.subscribe <- subscription
						h.subscribers[requestWrapper.listener] = true
					}
				} else {
					fmt.Println("Request doesn't contain subscription channel.")
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

		case requestWrapper := <-h.broadcast:
			fmt.Println(h.res + ": Broadcasting message.")

			// broadcasting a message to all connections. only the owner of the request doesn't receive this broadcast
			// because we send response message to the request
		for k, _ := range h.subscribers {
			if k != requestWrapper.listener {
				k <- requestWrapper.message
			}
		}

		case listener := <-h.subscribe:
			fmt.Println(h.res + ": Adding new listener to subscribers.")
			h.subscribers[listener] = true

		case listener := <-h.unsubscribe:
			fmt.Println(h.res + ": Removing a listener from subscribers.")
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
	h.broadcast = make(chan RequestWrapper)
	h.subscribe = make(chan chan Message)
	h.unsubscribe = make(chan chan Message)
	return
}

func createModelHolder(broadcastChannel chan RequestWrapper) (mh ModelHolder) {
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
