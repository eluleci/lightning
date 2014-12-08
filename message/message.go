package message

type Message struct {

	Rid                      int `json:"rid,omitempty"`
	Status                   int `json:"status,omitempty"`
	Res                      string `json:"res,omitempty"`
	Command                  string `json:"cmd,omitempty"`
	Parameters               string `json:"params,omitempty"`
	Body                     map[string]interface{} `json:"body,omitempty"`
}

// this struct is used for connections to send a message to a hub that they are not subscribed before
// this is created from a connection and sent to rootHub's inbox channel. rootHub finds the related hub and that hub
// executes the message inside the request. also, the connection receives a Subscription object from that hub
type RequestWrapper struct {
	Res       string
	Message   Message
	Listener  chan Message
	Subscribe chan Subscription
}

// this struct is used for notifying that a connection is subscribed to a hub. this can happen in two ways:
// 1: if connection sends a request wrapper to rootHub, rootHub finds the requested hub and gives the request. then the
// message is executed by the related hub and also the connection is subscribed to that hub
// 2: if a connection is subscribed to a list of resource, the connection is automatically subscribed to that list if a
// new item is created inside that list
type Subscription struct {
	Res                   string
	InboxChannel          chan RequestWrapper // inbox channel of hub to send message
	//	broadcastChannel      chan RequestWrapper // broadcast channel of hub to receive updates
	UnsubscriptionChannel chan RequestWrapper // unsubscription channel of hub for unsubscription
}
