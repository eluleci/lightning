package main

import (
	"fmt"
	"strings"
)

type Hub struct {
	res          string
	model        ModelHolder
	children     map[string]Hub
	subscribers  map[chan Message]chan Subscription
	inbox        chan RequestWrapper
	broadcast    chan RequestWrapper
	subscribe    chan RequestWrapper
	unsubscribe  chan RequestWrapper
}

func (h *Hub) run() {

	fmt.Println(h.res + ":  Started running.")

	if len(h.subscribers) > 0 {
		fmt.Println(h.res + ":  There are initial subscribers")
		h.printSubscribers()
	}

	for {
		select {
		case requestWrapper := <-h.inbox:
			fmt.Println(h.res+": Received message: ", requestWrapper.message)

			if requestWrapper.res == h.res {
				// if the resource of that message is this hub's resource
				fmt.Println(h.res + ": responsible of message.")

				if h.model.model["res"] != nil {
					// if model exists, forward message to model handler
					fmt.Println(h.res + ": Forwarding message to model handler")
					h.model.handle <- requestWrapper

				} else if (requestWrapper.message.Command == "post" && requestWrapper.message.Body["id"] != nil) {
					// if model doesn't exists, and if there is id in body, it means this is model initialisation message
					h.model = createModelHolder(h.res, h.broadcast)
					go h.model.run()
					h.model.handle <- requestWrapper

				} else if requestWrapper.message.Command == "post" && requestWrapper.message.Body["id"] == nil {
					// if model doesn't exists, if command is post, and if there is no id inside the message, it means
					// that this is a message for object creation under this domain
					h.createNewChild(requestWrapper)
					continue

				} else if requestWrapper.message.Command == "get" && len(h.children) > 0 {
					// if model doesn't exists, if command is get, and if there are children hubs, it means that this is
					// a domain hub and this message is a get message of the list
					fmt.Println(h.res + ": Getting list of the models in the resource " + h.res)
					list := make([]map[string]interface{}, len(h.children))
					callback := make(chan Message)

					// sending get messages and adding listener channel to all children as subscriber
					var getMessage Message
					getMessage.Command = "get"
					var rw RequestWrapper
					rw.message = getMessage
					rw.listener = callback
					for k, chlidrenHub := range h.children {
						rw.res = k
						chlidrenHub.addSubscription(requestWrapper)
						chlidrenHub.inbox <- rw
					}
					// receiving responses (receiving response is done after sending all messages for preventing being
					// blocked by a get message)
					for i := 0; i < len(h.children); i++ {
						response := <-callback
						//						fmt.Println(response.Body)
						list[i] = response.Body
					}
					var answer Message
					answer.Rid = requestWrapper.message.Rid
					answer.Res = h.res
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
					fmt.Println(h.res+": looking to connection list ", len(h.subscribers))
					//					if index := getIndex(requestWrapper.listener, h.subscribers); index == -1 {
					if _, exists := h.subscribers[requestWrapper.listener]; !exists {
						fmt.Println(h.res+": adding connection to subscription list ", requestWrapper.listener)
						go h.addSubscription(requestWrapper)
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
					hub = createHub(childRes, nil)
					go hub.run()
					h.children[childRes] = hub
					fmt.Println(h.res+": Created a hub for res: ", hub.res)
				}
				//   forward message to the children hub
				hub.inbox <- requestWrapper
			}

		case requestWrapper := <-h.broadcast:
			fmt.Println(h.res+": Broadcasting message. Number of subscribers: #", len(h.subscribers))

			// broadcasting a message to all connections. only the owner of the request doesn't receive this broadcast
			// because we send response message to the request
		for listenerChannel, _ := range h.subscribers {
			if listenerChannel != requestWrapper.listener {
				go h.checkAndSend(listenerChannel, requestWrapper.message)
			}
		}

		case requestWrapper := <-h.subscribe:

			// add the connection if it doesn't already exist
			//			if index := getIndex(listener, h.subscribers); index == -1 {
			if _, exists := h.subscribers[requestWrapper.listener]; !exists {
				//				h.subscribers = append(h.subscribers, listener)
				h.subscribers[requestWrapper.listener] = requestWrapper.subscribe
				fmt.Println(h.res+": Added new listener to subscribers. #", requestWrapper.listener)
				fmt.Println(h.res+": New subscribers count is #", len(h.subscribers))
			}

		case requestWrapper := <-h.unsubscribe:
			//			fmt.Println(h.res+": unsubscribing a listener from subscribers #", requestWrapper.listener)
			if _, exists := h.subscribers[requestWrapper.listener]; exists {
				delete(h.subscribers, requestWrapper.listener)
				fmt.Println(h.res+": removed a listener from subscribers. subscriptions remained #", len(h.subscribers))
			} else {
				fmt.Println(h.res + ": listener doesn't exists in subscriber list:")
				h.printSubscribers()
			}
		}
	}

}

func (h *Hub) addSubscription(requestWrapper RequestWrapper) {
	// add to subscribers if it doesn't already subscribed
	if _, exists := h.subscribers[requestWrapper.listener]; !exists {
		subscription := Subscription {
			h.res,
			h.inbox,
			h.unsubscribe,
		}
		//		go func() {
		requestWrapper.res = h.res
		h.subscribe <- requestWrapper
		//		}()
		requestWrapper.subscribe <- subscription
	}
}

func (h *Hub) checkAndSend(c chan Message, m Message) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("TODO: remove the channel from subscribers", r)
			//			h.unsubscribe <- c
		}
	}()
	c <- m
}

func (h *Hub) createNewChild(requestWrapper RequestWrapper) {

	generatedId := randSeq(32)
	generatedRes := requestWrapper.res + "/" + generatedId

	// copying subscribers of parent to pass to the newly created child hub
	subscribersCopy := make(map[chan Message]chan Subscription)
	for listenChannel, subscriptionChannel := range (h.subscribers) {
		subscribersCopy[listenChannel] = subscriptionChannel
	}

	// creating a child hub with initial subscribers
	hub := createHub(generatedRes, subscribersCopy)
	go hub.run()
	fmt.Println(h.res+": Created a new object for res: ", hub.res)

	// notifying connections that they are subscribed to a new hub
	for _, subscriptionChannel := range (subscribersCopy) {
		subscription := Subscription {hub.res, hub.inbox, hub.unsubscribe, }
		subscriptionChannel <- subscription
	}

	// adding new child to children hub
	h.children[hub.res] = hub

	// adding generated id and res to the request wrapper
	requestWrapper.message.Body["id"] = generatedId
	requestWrapper.message.Body["res"] = generatedRes

	// broadcasting the object creation. we need to create a new request wrapper because we are changing
	// the res of the request wrapper when sending it to newly created hub. but subscribers of this
	// domain will expect a broadcast message with the resource path of parent domain
	requestWrapperForBroadcast := requestWrapper
	requestWrapperForBroadcast.message.Rid = 0
	go func() {
		h.broadcast <- requestWrapperForBroadcast
	}()

	// sending request to the new hub with new res
	requestWrapper.message.Res = generatedRes
	requestWrapper.res = generatedRes
	hub.inbox <- requestWrapper
}

func createHub(res string, initialSubscribers map[chan Message]chan Subscription) (h Hub) {
	h.res = res
	h.children = make(map[string]Hub)
	h.inbox = make(chan RequestWrapper)
	h.broadcast = make(chan RequestWrapper)
	h.subscribe = make(chan RequestWrapper)
	h.unsubscribe = make(chan RequestWrapper)

	if initialSubscribers != nil {
		h.subscribers = initialSubscribers
	} else {
		h.subscribers = make(map[chan Message]chan Subscription, 0)
	}
	return
}

func createModelHolder(res string, broadcastChannel chan RequestWrapper) (mh ModelHolder) {
	mh.res = res
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

func (h *Hub) printSubscribers() {
	for k, _ := range h.subscribers {
		fmt.Println(k)
	}
}
