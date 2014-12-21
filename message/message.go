package message

type Message struct {

	Rid                      int `json:"rid,omitempty"`
	Res                      string `json:"res,omitempty"`
	Command                  string `json:"cmd,omitempty"`
	Headers                  map[string][]string `json:"headers,omitempty"`
	Parameters               map[string][]string `json:"parameters,omitempty"`
	Body                     map[string]interface{} `json:"body,omitempty"`
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
}

type RequestError struct {
	Code    int
	Message string
	Body    map[string]interface{}
}
