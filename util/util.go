package util

import (
	"math/rand"
	"github.com/eluleci/lightning/message"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func CreateErrorMessage(rid, status int, messageContent string) (m message.Message) {
	_ = messageContent
	m.Rid = rid
	m.Status = status
	return
}
