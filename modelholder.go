package main

import "fmt"

type ModelHolder struct {
	res              string
	model            map[string]interface{}
	handle           chan RequestWrapper
	broadcastChannel chan RequestWrapper
}

func (mh *ModelHolder) run() {

	for {
		select {
		case requestWrapper := <-mh.handle:
			if mh.model != nil {
				fmt.Println("MH("+mh.model["id"].(string)+"): received message: ", requestWrapper.message)
			}

			if mh.model == nil {
				// this is an object creation. so setting all body to the model
				mh.model = requestWrapper.message.Body
				//				fmt.Println("MH(" + mh.model["id"].(string) + "): Initialising the model.")
				answer := Message{}
				answer.Rid = requestWrapper.message.Rid
				answer.Res = mh.res
				answer.Status = 200
				answer.Body = make(map[string]interface{})
				answer.Body["id"] = mh.model["id"]
				answer.Body["res"] = mh.model["res"]
				requestWrapper.listener <- answer

			} else {
				// this is a regular message for the object
				//				fmt.Println("MH(" + mh.model["id"].(string) + "): Handling regular message.")
				message := requestWrapper.message

				if message.Command == "post" {
					// updating fields of the object
					for k, v := range message.Body {
						mh.model[k] = v
					}

					// returning response
					answer := Message{}
					answer.Rid = requestWrapper.message.Rid
					answer.Res = mh.res
					answer.Status = 200
					checkAndSend(requestWrapper.listener, answer)

					// broadcasting the updates
					requestWrapper.message.Rid = 0
					mh.broadcastChannel <- requestWrapper

				} else if message.Command == "get" {
					answer := Message{}
					answer.Rid = requestWrapper.message.Rid
					answer.Res = mh.res
					answer.Status = 200
					answer.Body = mh.model
					requestWrapper.listener <- answer
				}
			}
		}
	}
}
