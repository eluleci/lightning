package main

import (
	"fmt"
	"math/rand"
	"encoding/json"
)

type Lightning struct {

	model map[string]interface{}

	handle chan ConnectionRequest

	broadcast chan []byte

	// Handle the last message in the given connection
	bindNew chan Message
}

type ConnectionRequest struct {
	connection *connection
	message Message
}

func (l *Lightning) run() {
	for {
		select {
		case cm := <-l.handle:
			fmt.Println("Handling message...")

			str, _ := json.Marshal(cm.message)
			fmt.Println(string(str))

			if cm.connection == nil && cm.message.Body == nil {
				fmt.Println("creating a lightning object DOMAIN " + cm.message.Res)
				model := make(map[string]interface{})
				model["list"] = make([]interface{}, 0)
				continue
			}

			if cm.message.Command == "get" {
				fmt.Println("get message")
				fmt.Println("returning ")
				answerMessage := Response{cm.message.Rid, cm.message.Res, 200, l.model, ""}
				answer, _ := json.Marshal(answerMessage)
				cm.connection.send <- answer    // sending the answer back to the connection

			} else if cm.message.Command == "post" {

				objectId := cm.message.Body["id"]

				if objectId == nil {
					fmt.Println("This is list object. Adding new object to list " + cm.message.Res)

					generatedId := randSeq(32)

					// creating response message and sending it back
					body := make(map[string]interface{})
					body["id"] = generatedId
					answer, _ := json.Marshal(Response {cm.message.Rid, "", 200, body, ""})
					cm.connection.send <- answer            // sending the answer back to the connection

					// creating broadcast message and broadcasting it
					cm.message.Rid = 0
					cm.message.Body["id"] = generatedId
					broadcastMessage, _ := json.Marshal(cm.message)
					_ = broadcastMessage
					//				l.broadcast <- broadcastMessage            // broadcasting message to everyone in the hub

					// creating new hub for the newly created item
					bind <- ConnectionRequest{nil, cm.message}
					//					l.bindNew <- cm.message
				} else {
					fmt.Println("This is object. Putting the data into the model " + cm.message.Res)

					for k, v := range cm.message.Body {
						l.model[k] = v
					}
					//					l.model = cm.message.Body

					if cm.connection == nil {
						fmt.Println("There is no connection. This is the first creation of the object.")
					} else {
						fmt.Println("There is a connection. Sending response to the connection.")
						answer, _ := json.Marshal(Response {cm.message.Rid, "", 200, nil, ""})
						cm.connection.send <- answer            // sending the answer back to the connection
					}
				}

			}

			/*
						if cm.message.Command == "get" {
							fmt.Println("Read list message")
							answerMessage := Response{cm.message.Rid, cm.message.Res, 200, nil, ""}

							var list []interface{}
							if len(l.model) > 0 {
								list = l.People
							} else {
								list = make([]interface{}, 0)
							}
							answerMessage.Body = map[string] interface{}{
								"list": list,
							}
							answer, _ := json.Marshal(answerMessage)
							cm.connection.send <- answer    // sending the answer back to the connection

						} else if cm.message.Command == "post" {

							if cm.message.Ref != "" {
								// updating existing object
								fmt.Println("updating " + cm.message.Ref)

								profileId := cm.message.Ref
								for i, _ := range l.People {

									if l.People[i].Id == profileId {

										if cm.message.Body["name"] != nil {
											l.People[i].Name = cm.message.Body["name"].(string)
										}
										if cm.message.Body["avatar"] != nil {
											l.People[i].Avatar = cm.message.Body["avatar"].(string)
										}
										break
									}
								}
								answer, _ := json.Marshal(Response {cm.message.Rid, cm.message.Ref, 200, nil, ""})
								cm.connection.send <- answer    // sending the answer back to the connection

								cm.message.Rid = 0
								cm.message.Ref = profileId
								broadcastMessage, _ := json.Marshal(cm.message)
								l.broadcast <- broadcastMessage        // broadcasting message to everyone in the hub

							} else {
				// creating an object
				fmt.Println("creating new object")

				//					object := cm.message.Body
				cm.message.Body["id"] = randSeq(10)
				l.People = append(l.People, cm.message.Body)

				body := make(map[string]interface{})
				body["id"] = cm.message.Body["id"]

				answer, _ := json.Marshal(Response {cm.message.Rid, "", 200, body, ""})
				cm.connection.send <- answer            // sending the answer back to the connection
				cm.message.Rid = 0
				broadcastMessage, _ := json.Marshal(cm.message)
				l.broadcast <- broadcastMessage            // broadcasting message to everyone in the hub

				// creating new hub for the newly created item
				l.bindNew <- body["id"].(string)
				//}
			}*/
			//			fmt.Println("End of handle message")
		}
	}

}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
