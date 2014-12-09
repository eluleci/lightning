package adapter

import (
	"github.com/eluleci/lightning/message"
	"github.com/eluleci/lightning/config"
	"github.com/eluleci/lightning/util"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"bytes"
	"strconv"
)

const serverRoot = "https://api.parse.com/1/classes"
const parseAppId = "oxjtnRnmGUKyM9SFd1szSKzO9wKHGlgC6WgyRpq8"
const parseApiKey = "qJcOosDh8pywHdWKkVryWPoQFT1JMyoZYjMvnUul"

type RestAdapter struct {

}

func (adapter *RestAdapter) ExecuteGetRequest(requestWrapper message.RequestWrapper) (map[string]interface{}, []map[string]interface{}, *message.RequestError) {

	var targetUrl string
	object := make(map[string]interface{})

	// TODO use custom converter for different end points

	targetUrl = serverRoot+requestWrapper.Res
	util.Log("debug", "RestAdapter: Target url: "+targetUrl)

	response, requestErr := buildAndExecuteHttpRequest(requestWrapper, targetUrl)
	if requestErr != nil {
		util.Log("error", "RestAdapter: Error occured when making request. ")
		return nil, nil, &message.RequestError{500, "Error occured when making request.", nil}
	} else if response.StatusCode >= 200 && response.StatusCode < 300 {

		defer response.Body.Close()

		// getting data from response
		data, ioErr := ioutil.ReadAll(response.Body)
		if ioErr != nil {
			util.Log("error", "RestAdapter: Error occured when reading response from source.")
			return nil, nil, &message.RequestError{500, "Error occured when reading response from source.", nil}
		}

		var objectParseErr error
		// getting object out of the response
		object, objectParseErr = getJSONObjectFromResponse(data)
		if objectParseErr != nil {
			// if there is an error while getting object, try getting it as an array

			objects, listParseErr := getJSONArrayFromResponse(data)
			if listParseErr != nil {
				util.Log("error", "RestAdapter: Error occured when parsing data.")
				return nil, nil, &message.RequestError{500, "Error occured when parsing data that is received from server.", nil}
			} else {
				// if the list is successfully retrieved from the data, return the list
				util.Log("debug", "RestAdapter: Fetched list of objects from source. Length: "+strconv.Itoa(len(objects)))
				return nil, objects, nil
			}
		} else {
			// if the object is successfully constructed from the data, check that is it list wrapper or not

			if len(config.DefaultConfig.CollectionIdentifier) > 0 {
				// if there is an arrayIdentifier in configuration, check the list field exists in object or not

				if objectInArrayIdentifierField, exists := object[config.DefaultConfig.CollectionIdentifier]; exists {
					// if arrayIdentifier field exists in the object, try to extract an array from that field

					arrayData, _ := json.Marshal(objectInArrayIdentifierField)
					objects, listParseErr := getJSONArrayFromResponse(arrayData)
					if listParseErr != nil {
						// if the field couldn't be extracted as an array, return the object only
						util.Log("debug", "RestAdapter: Fetched an object from source.")
						return object, nil, nil
					}

					// if the list is successfully extracted from the object, return the list only
					util.Log("debug", "RestAdapter: Fetched list of objects from source. Length: "+strconv.Itoa(len(objects)))
					return nil, objects, nil
				}
			}
			util.Log("debug", "RestAdapter: Fetched an object from source.")
			return object, nil, nil
		}
	} else {
		err := generateRequestError(response)
		return nil, nil, &err
	}
	return nil, nil, nil
}

func (adapter *RestAdapter) ExecutePutRequest(requestWrapper message.RequestWrapper) (map[string]interface{}, *message.RequestError) {

	var targetUrl string

	// TODO use custom converter for different end points

	targetUrl = serverRoot+requestWrapper.Res

	response, requestErr := buildAndExecuteHttpRequest(requestWrapper, targetUrl)
	if requestErr != nil {
		util.Log("error", "RestAdapter: Error occured when making request. ")
		return nil, &message.RequestError{500, "Error occured when making request.", nil}
	} else if response.StatusCode >= 200 && response.StatusCode < 300 {
		defer response.Body.Close()

		// getting data from response
		data, ioErr := ioutil.ReadAll(response.Body)
		if ioErr != nil {
			util.Log("error", "RestAdapter: Error occured when reading response from source.")
			return nil, &message.RequestError{500, "Error occured when reading response from source.", nil}
		}

		// getting object out of the response
		object, objectParseErr := getJSONObjectFromResponse(data)
		if objectParseErr != nil {
			// if there is an error while getting object, try getting it as an array
			util.Log("error", "RestAdapter: Error occured when parsing data.")
			return nil, &message.RequestError{500, "Error occured when parsing data that is received from server.", nil}

		} else {
			util.Log("debug", "RestAdapter: Successfully parsed put response body.")
			return object, nil
		}
	} else {
		err := generateRequestError(response)
		return nil, &err
	}
	return nil, nil
}

func (adapter *RestAdapter) ExecutePostRequest(requestWrapper message.RequestWrapper) (map[string]interface{}, *message.RequestError) {

	var targetUrl string

	// TODO use custom converter for different end points

	targetUrl = serverRoot+requestWrapper.Res

	response, requestErr := buildAndExecuteHttpRequest(requestWrapper, targetUrl)
	if requestErr != nil {
		util.Log("error", "RestAdapter: Error occured when making request. ")
		return nil, &message.RequestError{500, "Error occured when making request.", nil}
	} else if response.StatusCode >= 200 && response.StatusCode < 300 {
		defer response.Body.Close()

		// getting data from response
		data, ioErr := ioutil.ReadAll(response.Body)
		if ioErr != nil {
			util.Log("error", "RestAdapter: Error occured when reading response from source.")
			return nil, &message.RequestError{500, "Error occured when reading response from source.", nil}
		}

		// getting object out of the response
		object, objectParseErr := getJSONObjectFromResponse(data)
		if objectParseErr != nil {
			// if there is an error while getting object, try getting it as an array
			util.Log("error", "RestAdapter: Error occured when parsing data.")
			return nil, &message.RequestError{500, "Error occured when parsing data that is received from server.", nil}

		} else {
			util.Log("debug", "RestAdapter: Response for post request is successful.")
			return object, nil
		}
	} else {
		err := generateRequestError(response)
		return nil, &err
	}
	return nil, nil
}

func (adapter *RestAdapter) ExecuteDeleteRequest(requestWrapper message.RequestWrapper) (map[string]interface{}, *message.RequestError) {

	var targetUrl string

	// TODO use custom converter for different end points

	targetUrl = serverRoot+requestWrapper.Res

	response, requestErr := buildAndExecuteHttpRequest(requestWrapper, targetUrl)
	if requestErr != nil {
		util.Log("error", "RestAdapter: Error occured when making request. ")
		return nil, &message.RequestError{500, "Error occured when making request.", nil}
	} else if response.StatusCode >= 200 && response.StatusCode < 300 {
		util.Log("debug", "RestAdapter: Response for delete request is successful.")
		return nil, nil
	} else {
		err := generateRequestError(response)
		return nil, &err
	}
	return nil, nil
}

func buildAndExecuteHttpRequest(requestWrapper message.RequestWrapper, url string) (resp *http.Response, err error) {
	client := &http.Client{}
	body, _ := json.Marshal(requestWrapper.Message.Body)
	request, _ := http.NewRequest(requestWrapper.Message.Command, url, bytes.NewBuffer(body))
	for k, v := range requestWrapper.Message.Headers {
		if len(v) > 0 {
			request.Header.Set(k, v)
		}
	}
	resp, err = client.Do(request)
	return
}

func getJSONObjectFromResponse(inputData []byte) (object map[string]interface{}, err error) {

	err = json.Unmarshal(inputData, &object)
	return
}

func getJSONArrayFromResponse(inputData []byte) (objects []map[string]interface{}, err error) {

	err = json.Unmarshal(inputData, &objects)
	return
}

func generateRequestError(response *http.Response) (err message.RequestError) {
	err.Code = response.StatusCode
	err.Message = "Request to source failed."
	data, ioErr := ioutil.ReadAll(response.Body)
	if ioErr == nil {
		errorBody, parseErr := getJSONObjectFromResponse(data)
		if parseErr == nil {
			err.Body = errorBody
		}
	}
	util.Log("debug", "RestAdapter: Request to source failed. "+strconv.Itoa(response.StatusCode))
	return
}
