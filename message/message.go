package message

type Message struct {

	Rid                      int `json:"rid,omitempty"`
	Res                      string `json:"res,omitempty"`
	Command                  string `json:"cmd,omitempty"`
	Headers                  map[string]string `json:"headers,omitempty"`
	Body                     map[string]interface{} `json:"body,omitempty"`
	Parameters               string `json:"params,omitempty"`
	Status                   int `json:"status,omitempty"` // used only in responses
}

type RequestWrapper struct {
	Res       string
	Message   Message
	Listener  chan Message
	Subscribe chan Subscription
}

type Subscription struct {
	Res                   string
	InboxChannel          chan RequestWrapper // inbox channel of hub to send message
	//	broadcastChannel      chan RequestWrapper // broadcast channel of hub to receive updates
	UnsubscriptionChannel chan RequestWrapper // unsubscription channel of hub for unsubscription
}

type RequestError struct {
	Code    int
	Message string
	Body    map[string]interface{}
}
