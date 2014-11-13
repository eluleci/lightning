package main

type Message struct {

	Rid                      int `json:"rid,omitempty"`
	Ref                      string `json:"ref,omitempty"`
	Command                  string `json:"cmd"`
	Parameters               string `json:"params,omitempty"`
	Body                map[string]interface{} `json:"body,omitempty"`
}

type Response struct {

	Rid                      int `json:"rid"`
	Ref                      string `json:"ref,omitempty"`
	Status                   int `json:"status"`
	Body                     interface{} `json:"body,omitempty"`
	Error                    string `json:"error,omitempty"`
}
