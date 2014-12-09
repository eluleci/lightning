package node

import (
	"github.com/eluleci/lightning/message"
	"github.com/eluleci/lightning/util"
	"encoding/json"
)

type ModelHolder struct {
	res              string
	model            map[string]interface{}
	handle           chan message.RequestWrapper
	broadcastChannel chan message.RequestWrapper
}

func (mh *ModelHolder) run() {

	for {
		select {
		case requestWrapper := <-mh.handle:
			if mh.model != nil {
				messageString, _ := json.Marshal(requestWrapper.Message)
				util.Log("debug", "MH-"+mh.model["::res"].(string)+": received message: "+string(messageString))
			}

			if mh.model == nil && requestWrapper.Message.Command == "initialise" {
				mh.model = requestWrapper.Message.Body
				util.Log("debug", "MH-"+mh.model["::res"].(string)+": Initialised the model.")

			} else if requestWrapper.Message.Command == "get" {

				answer := message.Message{}
				answer.Rid = requestWrapper.Message.Rid
				answer.Res = mh.res
				answer.Status = 200
				answer.Body = mh.model
				requestWrapper.Listener <- answer

			}

			/*if mh.model == nil {
				// this is an object creation. so setting all body to the model
				mh.model = requestWrapper.Message.Body
				//				fmt.Println("MH(" + mh.model["id"].(string) + "): Initialising the model.")
				answer := message.Message{}
				answer.Rid = requestWrapper.Message.Rid
				answer.Res = mh.res
				answer.Status = 200
				answer.Body = make(map[string]interface{})
				answer.Body["id"] = mh.model["id"]
				answer.Body["res"] = mh.model["res"]
				requestWrapper.Listener <- answer

			} else {
				// this is a regular message for the object
				//				fmt.Println("MH(" + mh.model["id"].(string) + "): Handling regular message.")
				msg := requestWrapper.Message

				if msg.Command == "post" {
					// updating fields of the object
					for k, v := range msg.Body {
						mh.model[k] = v
					}

					// returning response
					answer := message.Message{}
					answer.Rid = requestWrapper.Message.Rid
					answer.Res = mh.res
					answer.Status = 200
					checkAndSend(requestWrapper.Listener, answer)

					// broadcasting the updates
					requestWrapper.Message.Rid = 0
					mh.broadcastChannel <- requestWrapper

				} else if msg.Command == "get" {
					answer := message.Message{}
					answer.Rid = requestWrapper.Message.Rid
					answer.Res = mh.res
					answer.Status = 200
					answer.Body = mh.model
					requestWrapper.Listener <- answer
				}
			}*/
		}
	}

}
