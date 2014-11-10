package main

import (
	"encoding/json"
	"fmt"
)

type SessionManager struct {

	People []Profile

}

const (
	listSubscriptionId = "XVlBzgbAiC"
)

func (sessionManager *SessionManager) ProcessMessage(receivedMessage []byte) (answer []byte, broadcastMessage []byte) {

	var message Message
	var responseObject Response

	fmt.Println("Processing message...")
	fmt.Println(string(receivedMessage))

	err := json.Unmarshal([]byte(string(receivedMessage[:])), &message)
	if err != nil {
		fmt.Println("error while parsing message: ", err)
		responseObject = Response {message.Rid, "", 400, nil, "error while parsing message"}
		response, _ := json.Marshal(responseObject)
		answer = []byte(response)
		broadcastMessage = nil
		return
	}

	if message.Command == "read" {
		fmt.Println("Read list message")
		answerMessage := Response{message.Rid, listSubscriptionId, 200, nil, ""}
		answerMessage.Body = map[string] interface{}{
			"list": GetCurrentState(),
		}
		answer, _ = json.Marshal(answerMessage)
		broadcastMessage = nil

	} else if message.Command == "create" {
		//		sessionManager.People = ApplyMessage(message, sessionManager.People)

		var profile Profile
		profile.Id = randSeq(10)
		profile.Name = message.Body["name"].(string)
		profile.Avatar = message.Body["avatar"].(string)
		sessionManager.People = append(sessionManager.People, profile)

		body := make(map[string]interface{})
		body["id"] = profile.Id

		answer, _ = json.Marshal(Response {message.Rid, "", 200, body, ""})
		message.Rid = 0
		message.Body["id"] = profile.Id
		message.Subscription = listSubscriptionId
		broadcastMessage, _ = json.Marshal(message)
	} else if message.Command == "update" {

		fmt.Println("updating " + message.Body["id"].(string))

		var profile Profile
		profile.Id = message.Body["id"].(string)
		profile.Name = message.Body["name"].(string)
		profile.Avatar = message.Body["avatar"].(string)

		for i, _ := range sessionManager.People {

			if sessionManager.People[i].Id == profile.Id {

				if profile.Name != "" {
					sessionManager.People[i].Name = profile.Name
				}
				if profile.Avatar != "" {
					sessionManager.People[i].Avatar = profile.Avatar
				}
				break
			}
		}
		answer, _ = json.Marshal(Response {message.Rid, "", 200, nil, ""})

		message.Rid = 0
		message.Subscription = profile.Id
		broadcastMessage, _ = json.Marshal(message)

	}

	fmt.Println(string(answer))
	fmt.Println("Returned response")
	return
}

func (sessionManager *SessionManager) GetCurrentState() ([]Profile) {

	if len(sessionManager.People) > 0 {
		array := sessionManager.People
		return array
	} else {
		array := make([]Profile, 0)
		return array
	}

}
