package main

import "math/rand"

func ApplyMessage(message Message, people []Profile) []Profile {

	if message.Command == "create" {
		// creating profile object

		var profile Profile
		profile.Id = randSeq(10)
		profile.Name = message.Body["name"].(string)
		profile.Avatar = message.Body["avatar"].(string)
		people = append(people, profile)
		return people
	} else if message.Command == "update" {

	}
	return people

	// detect message action
	// if message command == create && params.type == person
	//		-> create new person and add to array
	// if message command == create && params.type == moment
	//		-> find the person from the list
	//		-> add the moment
	// if message command == update && params.type == person

}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}



