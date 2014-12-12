package adapter

import "github.com/eluleci/lightning/message"

type Adapter interface {
	ExecuteGetRequest(requestWrapper message.RequestWrapper) (map[string]interface{}, []map[string]interface{}, *message.RequestError)
	ExecutePutRequest(requestWrapper message.RequestWrapper) (map[string]interface{}, *message.RequestError)
	ExecutePostRequest(requestWrapper message.RequestWrapper) (map[string]interface{}, *message.RequestError)
	ExecuteDeleteRequest(requestWrapper message.RequestWrapper) (map[string]interface{}, *message.RequestError)
}
