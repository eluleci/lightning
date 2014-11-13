package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

type Lightning struct {

	People []interface{}

	handle chan ConnectionBinding

	broadcast chan []byte

	// Handle the last message in the given connection
	bindNew chan string
}

type ConnectionBinding struct {
	reference string
	connection *connection
	message Message
}

func (l *Lightning) run() {
	for {
		select {
		case cm := <-l.handle:
			fmt.Println("Handling message...")

			if cm.message.Command == "get" {
				fmt.Println("Read list message")
				answerMessage := Response{cm.message.Rid, cm.message.Ref, 200, nil, ""}

				var list []interface{}
				if len(l.People) > 0 {
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

				/*if cm.message.Ref != "" {
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

				} else {*/
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
			}
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
