package adapter
/*

import (
	"github.com/eluleci/lightning/message"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
	"fmt"
	"strings"
)

type MongoAdapter struct {
}

func (adapter MongoAdapter) ExecuteGetRequest(requestWrapper message.RequestWrapper) (map[string]interface{}, []map[string]interface{}, *message.RequestError) {

	return nil, nil, nil
}

func (adapter MongoAdapter) ExecutePutRequest(requestWrapper message.RequestWrapper) (map[string]interface{}, *message.RequestError) {

	return nil, nil
}

func (adapter MongoAdapter) ExecutePostRequest(requestWrapper message.RequestWrapper) (map[string]interface{}, *message.RequestError) {

	mongoSession, err := mgo.Dial("localhost")
	if err != nil {
		return nil, nil
	}
	database := mongoSession.DB("thunderdock")

	//	className := requestWrapper.Message.Body["::class"]
	className := strings.Trim(requestWrapper.Res, "/")
	if len(className) == 0 {
		fmt.Printf("There must be a class in body.")
		return nil, nil
	}
	collection := database.C(className)

	err = collection.Insert(requestWrapper.Message.Body)
	if err != nil {
		fmt.Printf("Can't insert document: %v\n", err)
	}
	return nil, nil
}

func (adapter MongoAdapter) ExecuteDeleteRequest(requestWrapper message.RequestWrapper) (map[string]interface{}, *message.RequestError) {

	return nil, nil
}
*/
