package node

import (
	"strings"
	"encoding/json"
	"strconv"
	"github.com/eluleci/lightning/message"
	"github.com/eluleci/lightning/adapter"
	"github.com/eluleci/lightning/util"
	"github.com/eluleci/lightning/config"
)

type Hub struct {
	res               string
	model             ModelHolder
	children          map[string]Hub
	subscribers       map[chan message.Message]chan message.Subscription
	Inbox             chan message.RequestWrapper
	broadcast         chan message.RequestWrapper
	parentInbox       chan message.RequestWrapper
	unsubscribe       chan message.RequestWrapper
	adapter           adapter.RestAdapter
}

func (h *Hub) Run() {

	util.Log("debug", h.res+":  Started running.")

	if len(h.subscribers) > 0 {
		util.Log("debug", h.res+": Hub has initial subscribers #"+strconv.Itoa(len(h.subscribers)))
	}

	for {
		select {
		case requestWrapper := <-h.Inbox:

			messageString, _ := json.Marshal(requestWrapper.Message)
			util.Log("debug", h.res+": Received message: "+string(messageString))

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
					h.executeDeleteOnAdapter(requestWrapper)

				} else if requestWrapper.Message.Command == "::initialise" {
					// this is an object initialisation message. this hub is responsible of an existing object that is
					// provided in the request wrapper
					h.initialiseModel(requestWrapper)
				} else if requestWrapper.Message.Command == "::deleteChild" {
					// this is a message from child hub for its' deletion. when a parent hub receives this message, it
					// means that the child hub is deleted explicitly.

					childRes := requestWrapper.Message.Body["::res"].(string)
					if _, exists := h.children[childRes]; exists {

						// send broadcast message of the object deletion
						requestWrapper.Message.Command = "delete"
						requestWrapper.Message.Res = h.res
						go func() {
							h.broadcast <- requestWrapper
						}()

						// delete the child hub
						delete(h.children, childRes)
					}
				}

			} else {
				// if the resource belongs to a children hub
				_, childRes := getChildRes(requestWrapper.Res, h.res)

				hub, exists := h.children[childRes]
				if !exists {
					//   if children doesn't exists, create children hub for the res
					hub = CreateHub(childRes, nil, h.Inbox)
					go hub.Run()
					h.children[childRes] = hub
				}
				//   forward message to the children hub
				hub.Inbox <- requestWrapper
			}

		case requestWrapper := <-h.broadcast:
			util.Log("debug", h.res+": Broadcasting message. Number of subscribers: #"+strconv.Itoa(len(h.subscribers)))

			// broadcasting a message to all connections. only the owner of the request doesn't receive this broadcast
			// because we send 'response message' to the request owner
		for listenerChannel, _ := range h.subscribers {
			if listenerChannel != requestWrapper.Listener {
				go h.checkAndSend(listenerChannel, requestWrapper.Message)
			}
		}

		case requestWrapper := <-h.unsubscribe:

			// remove listener from subscribers if it is in subscribers list
			if _, exists := h.subscribers[requestWrapper.Listener]; exists {
				delete(h.subscribers, requestWrapper.Listener)
				util.Log("debug", h.res+": Removed a listener from subscribers. Subscriptions remained: #"+strconv.Itoa(len(h.subscribers)))
			} else {
				util.Log("debug", h.res+": The channel to remove doesn't exists in subscriber list.")
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
		util.Log("error", h.res+"Error occured when getting data from adapter. ")
		// TODO get more specific error from the adapter
		answer.Status = 404

	} else if object != nil {
		// if object is not null, it means that this is the object that this hub is responsible of
		util.Log("debug", h.res+": Received one object from adapter with id "+object[config.DefaultConfig.ObjectIdentifier].(string))

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
		util.Log("debug", h.res+": Received list of objects from adapter. Length: "+strconv.Itoa(len(objectArray)))

		// creating a new child hub and  adding it to children hub list
		for _, objectData := range (objectArray) {

			// generating res of the object: parentRes/objectId
			childRes := h.res + "/" + objectData[config.DefaultConfig.ObjectIdentifier].(string)
			objectData["::res"] = childRes

			if existingChild, exists := h.children[childRes]; !exists {
				childHub := h.generateChild(childRes, objectData)
				h.children[childHub.res] = childHub
			} else {
				// adding the listener to childs subscribers
				existingChild.addSubscription(requestWrapper)
				// TODO decide to give the fresh data to child hub or not
				util.Log("debug", h.res+": Child already exists for res "+childRes)
			}
		}

		answer.Status = 200
		answer.Body = make(map[string]interface{})
		answer.Body["::list"] = objectArray
	} else {
		util.Log("debug", h.res+": Received object or list from adapter failed.")
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
		util.Log("error", "Error occured when putting data with adapter. ")
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
		util.Log("error", "Error occured when posting data with adapter. ")
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

func (h *Hub) executeDeleteOnAdapter(requestWrapper message.RequestWrapper) {

	var answer message.Message
	answer.Rid = requestWrapper.Message.Rid
	answer.Res = h.res

	_, adapterErr := h.adapter.ExecuteDeleteRequest(requestWrapper)
	if adapterErr != nil {
		util.Log("error", "Error occured when posting data with adapter. ")
		// TODO get more specific error from the adapter
		answer.Status = 404
	} else {
		// if there is no error, it means that the object is deleted successfully
		answer.Status = 200

		// send broadcast message of the object deletion
		requestWrapper.Message.Rid = 0
		go func() {
			h.broadcast <- requestWrapper
		}()

		var deleteRequest message.RequestWrapper
		deleteRequest.Message.Command = "::deleteChild"
		deleteRequest.Res = getParentRes(h.res)
		deleteRequest.Message.Body = make(map[string]interface{})
		deleteRequest.Message.Body["::res"] = h.res
		deleteRequest.Listener = requestWrapper.Listener       // for not sending push message from parent to connection
		h.parentInbox <- deleteRequest
	}

	// sending result of DELETE message
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
		util.Log("debug", h.res+": Added new listener to subscribers. New size: #"+strconv.Itoa(len(h.subscribers)))
	}
}

func (h *Hub) checkAndSend(c chan message.Message, m message.Message) {
	defer func() {
		if r := recover(); r != nil {
			util.Log("debug", h.res+"Trying to send on closed channel. Removing channel from subscribers.")
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
	hub := CreateHub(objectRes, subscribersCopy, h.Inbox)
	go hub.Run()
	util.Log("debug", h.res+": Created a new child for res: "+hub.res+", with subscribers #"+strconv.Itoa(len(h.subscribers)))

	// creating model holder if PersistInMemory enabled
	if config.DefaultConfig.PersistInMemory {
		initialiseRequest := createInitialiseRequest(objectData, objectRes)
		hub.Inbox <- initialiseRequest
	}
	return hub
}

func createInitialiseRequest(objectData map[string]interface{}, objectRes string) message.RequestWrapper {

	var initialiseMessage message.Message
	initialiseMessage.Command = "::initialise"
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
	hub := CreateHub(generatedRes, subscribersCopy, h.Inbox)
	go hub.Run()

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

func CreateHub(res string, initialSubscribers map[chan message.Message]chan message.Subscription, parentInbox chan message.RequestWrapper) (h Hub) {
	h.res = res
	h.children = make(map[string]Hub)
	h.Inbox = make(chan message.RequestWrapper)
	h.broadcast = make(chan message.RequestWrapper)
	h.parentInbox = parentInbox
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

func getParentRes(res string) (path string) {
	res = strings.Trim(res, "/")
	li := strings.LastIndex(res, "/")
	path = "/"+res[:li]
	return
}
