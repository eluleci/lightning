package node

import (
	"fmt"
	"strings"
	"github.com/eluleci/lightning/message"
	"github.com/eluleci/lightning/adapter"
	"github.com/eluleci/lightning/util"
	"github.com/eluleci/lightning/config"
)

type Hub struct {
	res             string
	model           ModelHolder
	children        map[string]Hub
	subscribers     map[chan message.Message]chan message.Subscription
	Inbox           chan message.RequestWrapper
	broadcast       chan message.RequestWrapper
	subscribe       chan message.RequestWrapper
	unsubscribe     chan message.RequestWrapper
	adapter         adapter.RestAdapter
}

func (h *Hub) Run() {

	fmt.Println(h.res + ":  Started running.")

	if len(h.subscribers) > 0 {
		fmt.Println(h.res + ": Hub has initial subscribers.")
		h.printSubscribers()
	}

	for {
		select {
		case requestWrapper := <-h.Inbox:
			fmt.Println(h.res+": Received message: ", requestWrapper.Message)

			if requestWrapper.Res == h.res {
				// if the resource of the message is this hub's resource

				// if there is a subscription channel inside the request, subscribed the request sender
				// we need to subscribe the channel before we continue because there may be children hub creation
				// afterwords and we need to give all subscriptions of this hub to it's children
				if requestWrapper.Subscribe != nil {
					h.addSubscription(requestWrapper)
				}

				if requestWrapper.Message.Command == "get" {

					if config.DefaultConfig.PersistInMemory && h.model.model["::res"] != nil {
						// if persisting in memory and if the model exists, it means we already fetched data before.
						// so forward request to model holder
						h.model.handle <- requestWrapper

					} else {
						// if there is no model, and if there is adapter, get the
						// data from the adapter first.
						h.executeGetOnAdapter(requestWrapper)
					}
					/* else if len(h.children) > 0 {
	// if command is 'get', if model doesn't exists, and if there are children hubs, it means that
	// this is  domain hub and this message is a get message for the list
	fmt.Println(h.res + ": Returning list of models.")
	h.returnChildListToRequest(requestWrapper)

}*/

				} else if requestWrapper.Message.Command == "put" {
					// it is an update request.
					if config.DefaultConfig.PersistInMemory && h.model.model["::res"] != nil {
						// if persisting in memory and if the model exists, it means we already fetched data before.
						// so forward request to model holder
						h.model.handle <- requestWrapper
					} else {
						// if there is no model, and if there is adapter, execute the request from adapter directly
						h.executePutOnAdapter(requestWrapper)
					}

				}  else if requestWrapper.Message.Command == "post" {
					// it is an object creation message under this domain
					h.executePostOnAdapter(requestWrapper)

				}  else if requestWrapper.Message.Command == "delete" {
					// it is an object creation message under this domain
					// TODO: execute request with adapter
					// TODO: remove the hub from its' parent if result is successful
					// TODO: return the result to listener
					// TODO: broadcast the object deletion (must be done by parent)

				} else if requestWrapper.Message.Command == "initialise" {
					// this is an object initialisation message. this hub is responsible of an existing object that is
					// provided in the request wrapper
					h.initialiseModel(requestWrapper)
				}

				/*if h.model.model["res"] != nil {
					// if model exists, forward message to model handler
					go func() {
						h.model.handle <- requestWrapper
					}()

				} else if (requestWrapper.Message.Command == "post" && requestWrapper.Message.Body["id"] != nil) {
					// if model doesn't exists, and if there is id in body, it means this is model initialisation message
					h.model = createModelHolder(h.res, h.broadcast)
					go h.model.run()
					h.model.handle <- requestWrapper

				} else if requestWrapper.Message.Command == "post" && requestWrapper.Message.Body["id"] == nil {
					// if model doesn't exists, if command is post, and if there is no id inside the message, it means
					// that this is a message for object creation under this domain
					h.createNewChild(requestWrapper)
					continue    // calling continue not to subscribe the request at the end of the if statement

				} else if requestWrapper.Message.Command == "get" && len(h.children) > 0 {
					// if model doesn't exists, if command is get, and if there are children hubs, it means that this is
					// a domain hub and this message is a get message of the list
					fmt.Println(h.res + ": Returning list of models.")
					h.returnChildListToRequest(requestWrapper)
				} else {
					// if model doesn't exists, and if there is no children hub, it means that the resource doesn't exists
					requestWrapper.Listener <- util.CreateErrorMessage(requestWrapper.Message.Rid, 404, "Not found.")
					continue    // calling continue not to subscribe the request at the end of the if statement
				}*/



			} else {
				// if the resource belongs to a children hub
				_, childRes := getChildRes(requestWrapper.Res, h.res)

				hub, exists := h.children[childRes]
				if !exists {
					//   if children doesn't exists, create children hub for the res
					hub = CreateHub(childRes, nil)
					go hub.Run()
					h.children[childRes] = hub
				}
				//   forward message to the children hub
				hub.Inbox <- requestWrapper
			}

		case requestWrapper := <-h.broadcast:
			fmt.Println(h.res+": Broadcasting message. Number of subscribers: #", len(h.subscribers))

			// broadcasting a message to all connections. only the owner of the request doesn't receive this broadcast
			// because we send 'response message' to the request owner
		for listenerChannel, _ := range h.subscribers {
			if listenerChannel != requestWrapper.Listener {
				go checkAndSend(listenerChannel, requestWrapper.Message)
			}
		}

		case requestWrapper := <-h.subscribe:

			// add the connection if it is not already in subscribers list
			if _, exists := h.subscribers[requestWrapper.Listener]; !exists {
				//				h.subscribers = append(h.subscribers, listener)
				h.subscribers[requestWrapper.Listener] = requestWrapper.Subscribe
				fmt.Println(h.res+": Added new listener to subscribers. New size: #", len(h.subscribers))
			}

		case requestWrapper := <-h.unsubscribe:

			// remove listener from subscribers if it is in subscribers list
			if _, exists := h.subscribers[requestWrapper.Listener]; exists {
				delete(h.subscribers, requestWrapper.Listener)
				fmt.Println(h.res+": Removed a listener from subscribers. Subscriptions remained: #", len(h.subscribers))
			} else {
				fmt.Println(h.res + ": Listener doesn't exists in subscriber list:")
				h.printSubscribers()
			}
		}
	}

}

func (h *Hub) executeGetOnAdapter(requestWrapper message.RequestWrapper) {

	var answer message.Message
	answer.Rid = requestWrapper.Message.Rid
	answer.Res = h.res

	object, objectArray, adapterErr := h.adapter.ExecuteGetRequest(requestWrapper)
	if adapterErr != nil {
		fmt.Printf("Error occured when getting data from adapter. ", adapterErr)
		// TODO get more specific error from the adapter
		answer.Status = 404

	} else if object != nil {
		// if object is not null, it means that this is the object that this hub is responsible of

		// adding a new field to object body for subscription purposes
		object["::res"] = h.res

		answer.Status = 200
		answer.Body = object

		// creating model holder if PersistInMemory enabled
		if config.DefaultConfig.PersistInMemory {
			initialiseRequest := createInitialiseRequest(object, h.res)
			h.initialiseModel(initialiseRequest)
		}

	} else if objectArray != nil {
		// if object array is not null, it means that this hub is responsible of the collections of
		// of these objects. so we create a new hub for each object in the list and return the
		// result to listener

		// creating a new child hub and  adding it to children hub list
		for _, objectData := range (objectArray) {

			// generating res of the object: parentRes/objectId
			childRes := h.res + "/" + objectData[config.DefaultConfig.ObjectIdentifier].(string)
			objectData["::res"] = childRes

			if _, exists := h.children[childRes]; !exists {
				childHub := h.generateChild(childRes, objectData)
				h.children[childHub.res] = childHub
			} else {
				// TODO decide to give the fresh data to child hub or not
				fmt.Println("Child already exists")
			}
		}

		answer.Status = 200
		answer.Body = make(map[string]interface{})
		answer.Body["::list"] = objectArray
	} else {
		answer.Status = 404
	}

	// sending result of GET message
	requestWrapper.Listener <- answer
}

func (h *Hub) executePutOnAdapter(requestWrapper message.RequestWrapper) {

	var answer message.Message
	answer.Rid = requestWrapper.Message.Rid
	answer.Res = h.res

	response, adapterErr := h.adapter.ExecutePutRequest(requestWrapper)
	if adapterErr != nil {
		fmt.Printf("Error occured when putting data with adapter. ", adapterErr)
		// TODO get more specific error from the adapter
		answer.Status = 404

	} else if response != nil {

		answer.Status = 200
		answer.Body = response
		answer.Body["updatedAt"] = response["updatedAt"]

		// TODO: update the model holder if exists

		requestWrapper.Message.Rid = 0
		requestWrapper.Message.Body["updatedAt"] = response["updatedAt"]
		go func() {
			h.broadcast <- requestWrapper
		}()

	} else {
		answer.Status = 404
	}

	// sending result of GET message
	requestWrapper.Listener <- answer
}

func (h *Hub) executePostOnAdapter(requestWrapper message.RequestWrapper) {

	var answer message.Message
	answer.Rid = requestWrapper.Message.Rid
	answer.Res = h.res

	response, adapterErr := h.adapter.ExecutePostRequest(requestWrapper)
	if adapterErr != nil {
		fmt.Printf("Error occured when posting data with adapter. ", adapterErr)
		// TODO get more specific error from the adapter
		answer.Status = 404

	} else if response != nil {

		objectData := requestWrapper.Message.Body

		// adding a new field 'res' to object body for subscription purposes
		objectRes := h.res + "/" + response[config.DefaultConfig.ObjectIdentifier].(string)
		objectData["::res"] = objectRes
		objectData["createdAt"] = response["createdAt"]
		response["::res"] = objectRes

		answer.Status = 200
		answer.Res = objectRes
		answer.Body = response

		// generating new child hub for newly created object
		childHub := h.generateChild(objectRes, objectData)
		h.children[childHub.res] = childHub

		requestWrapper.Message.Rid = 0
		requestWrapper.Message.Body = objectData
		go func() {
			h.broadcast <- requestWrapper
		}()

	} else {
		answer.Status = 404
	}

	// sending result of GET message
	requestWrapper.Listener <- answer
}

func (h *Hub) initialiseModel(requestWrapper message.RequestWrapper) {

	h.model = createModelHolder(h.res, h.broadcast)
	go h.model.run()
	h.model.handle <- requestWrapper
}

func (h *Hub) addSubscription(requestWrapper message.RequestWrapper) {
	defer func() {
		if r := recover(); r != nil {
			// the subscribe channel may be closed. catching the panic
		}
	}()

	// add the connection if it is not already in subscribers list
	if _, exists := h.subscribers[requestWrapper.Listener]; !exists {
		subscription := message.Subscription {
			h.res,
			h.Inbox,
			h.unsubscribe,
		}
		requestWrapper.Subscribe <- subscription
		h.subscribers[requestWrapper.Listener] = requestWrapper.Subscribe
		fmt.Println(h.res+": Added new listener to subscribers. New size: #", len(h.subscribers))
	}
}

func checkAndSend(c chan message.Message, m message.Message) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("TODO: remove the channel from subscribers", r)
			//			h.unsubscribe <- c
		}
	}()
	c <- m
}

func (h *Hub) generateChild(objectRes string, objectData map[string]interface{}) Hub {

	// copying subscribers of parent to pass to the newly created child hub
	subscribersCopy := make(map[chan message.Message]chan message.Subscription)
	for listenChannel, subscriptionChannel := range (h.subscribers) {
		subscribersCopy[listenChannel] = subscriptionChannel
	}

	// creating a child hub with initial subscribers
	hub := CreateHub(objectRes, subscribersCopy)
	go hub.Run()
	fmt.Println(h.res+": Created a new object for res: ", hub.res)

	initialiseRequest := createInitialiseRequest(objectData, objectRes)
	hub.Inbox <- initialiseRequest

	return hub
}

func createInitialiseRequest(objectData map[string]interface{}, objectRes string) message.RequestWrapper {

	var initialiseMessage message.Message
	initialiseMessage.Command = "initialise"
	initialiseMessage.Body = objectData

	var initialiseRequest message.RequestWrapper
	initialiseRequest.Res = objectRes
	initialiseRequest.Message = initialiseMessage
	return initialiseRequest
}

/**
 * Creates a new object with the given data. Works independent fron adapter
 */
func (h *Hub) createNewChild(requestWrapper message.RequestWrapper) {

	generatedId := util.RandSeq(32)
	generatedRes := requestWrapper.Res + "/" + generatedId

	// copying subscribers of parent to pass to the newly created child hub
	subscribersCopy := make(map[chan message.Message]chan message.Subscription)
	for listenChannel, subscriptionChannel := range (h.subscribers) {
		subscribersCopy[listenChannel] = subscriptionChannel
	}

	// creating a child hub with initial subscribers
	hub := CreateHub(generatedRes, subscribersCopy)
	go hub.Run()
	fmt.Println(h.res+": Created a new object for res: ", hub.res)

	// adding new child to children hub
	h.children[hub.res] = hub

	// adding generated id and res to the request wrapper
	requestWrapper.Message.Body["id"] = generatedId
	requestWrapper.Message.Body["res"] = generatedRes

	// broadcasting the object creation. we need to create a new request wrapper because we are changing
	// the res of the request wrapper when sending it to newly created hub. but subscribers of this
	// domain will expect a broadcast message with the resource path of parent domain
	requestWrapperForBroadcast := requestWrapper
	requestWrapperForBroadcast.Message.Rid = 0
	go func() {
		h.broadcast <- requestWrapperForBroadcast
	}()

	// sending request to the new hub with new res
	requestWrapper.Message.Res = generatedRes
	requestWrapper.Res = generatedRes
	hub.Inbox <- requestWrapper
}

func (h *Hub) returnChildListToRequest(requestWrapper message.RequestWrapper) {
	list := make([]map[string]interface{}, len(h.children))
	callback := make(chan message.Message)

	// sending get messages and adding listener channel to all children as subscriber
	var getMessage message.Message
	getMessage.Command = "get"
	var rw message.RequestWrapper
	rw.Message = getMessage
	rw.Listener = callback
	for k, chlidrenHub := range h.children {
		rw.Res = k
		chlidrenHub.addSubscription(requestWrapper)
		chlidrenHub.Inbox <- rw
	}
	// receiving responses (receiving response is done after sending all messages for preventing being
	// blocked by a get message)
	for i := 0; i < len(h.children); i++ {
		response := <-callback
		//						fmt.Println(response.Body)
		list[i] = response.Body
	}
	var answer message.Message
	answer.Rid = requestWrapper.Message.Rid
	answer.Res = h.res
	answer.Status = 200
	answer.Body = make(map[string]interface{})
	answer.Body["list"] = list
	requestWrapper.Listener <- answer
}

func CreateHub(res string, initialSubscribers map[chan message.Message]chan message.Subscription) (h Hub) {
	h.res = res
	h.children = make(map[string]Hub)
	h.Inbox = make(chan message.RequestWrapper)
	h.broadcast = make(chan message.RequestWrapper)
	h.subscribe = make(chan message.RequestWrapper)
	h.unsubscribe = make(chan message.RequestWrapper)
	h.adapter = adapter.RestAdapter{}

	if initialSubscribers != nil {
		h.subscribers = initialSubscribers

		// notifying connections that they are subscribed to a new hub
		for _, subscriptionChannel := range (initialSubscribers) {
			subscription := message.Subscription {h.res, h.Inbox, h.unsubscribe, }
			subscriptionChannel <- subscription
		}
	} else {
		h.subscribers = make(map[chan message.Message]chan message.Subscription, 0)
	}
	return
}

func createModelHolder(res string, broadcastChannel chan message.RequestWrapper) (mh ModelHolder) {
	mh.res = res
	mh.handle = make(chan message.RequestWrapper)
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
