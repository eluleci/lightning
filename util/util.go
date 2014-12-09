package util

import (
	"math/rand"
	"github.com/eluleci/lightning/message"
	"fmt"
	"os"
	"log"
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

var logFile *os.File

func Log(level, message string) {

	go func() {
		var err error
		logFile, err = os.OpenFile("log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("error opening file: %v", err)
		}
		log.SetOutput(logFile)
		log.SetFlags(log.Lmicroseconds)
		defer logFile.Close()

		log.Println(message)
	}()
}
