package main

type Message struct {

	Rid                      int `json:"rid,omitempty"`
	Status                   int `json:"status"`
	Res                      string `json:"res,omitempty"`
	Command                  string `json:"cmd,omitempty"`
	Parameters               string `json:"params,omitempty"`
	Body                map[string]interface{} `json:"body,omitempty"`
}
