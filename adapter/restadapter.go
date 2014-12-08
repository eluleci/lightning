package adapter

import (
	"github.com/eluleci/lightning/message"
	"github.com/eluleci/lightning/config"
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"errors"
	"bytes"
)

const serverRoot = "https://api.parse.com/1/classes"
const parseAppId = "oxjtnRnmGUKyM9SFd1szSKzO9wKHGlgC6WgyRpq8"
const parseApiKey = "qJcOosDh8pywHdWKkVryWPoQFT1JMyoZYjMvnUul"

type RestAdapter struct {

}

func (adapter *RestAdapter) ExecuteGetRequest(requestWrapper message.RequestWrapper) (map[string]interface{}, []map[string]interface{}, error) {

	var targetUrl string
	object := make(map[string]interface{})

	// TODO use custom converter for different end points

	targetUrl = serverRoot+requestWrapper.Res
	fmt.Println("targetUrl: ", targetUrl)

	response, requestErr := buildAndExecuteHttpRequest(requestWrapper, targetUrl)
	if requestErr != nil {
		fmt.Printf("RestAdapter: Error occured when making request. ", requestErr)
		return nil, nil, requestErr
	} else if response.StatusCode == 404 {
		err := errors.New("Not found")
		return nil, nil, err
	} else {
		defer response.Body.Close()

		// getting data from response
		data, ioErr := ioutil.ReadAll(response.Body)
		if ioErr != nil {
			fmt.Printf("RestAdapter: Error occured when reading response. ", ioErr)
			return nil, nil, ioErr
		}

		var objectParseErr error
		// getting object out of the response
		object, objectParseErr = getJSONObjectFromResponse(data)
		if objectParseErr != nil {
			// if there is an error while getting object, try getting it as an array

			objects, listParseErr := getJSONArrayFromResponse(data)
			if listParseErr != nil {
				fmt.Printf("RestAdapter: Error occured when parsing data.")
				return nil, nil, listParseErr
			} else {
				// if the list is successfully retrieved from the data, return the list
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
						return object, nil, nil
					}

					// if the list is successfully extracted from the object, return the list only
					return nil, objects, nil
				}
			}
			return object, nil, nil
		}
	}
	return nil, nil, nil
}

func (adapter *RestAdapter) ExecutePutRequest(requestWrapper message.RequestWrapper) (map[string]interface{}, error) {

	var targetUrl string

	// TODO use custom converter for different end points

	targetUrl = serverRoot+requestWrapper.Res
	fmt.Println("targetUrl: ", targetUrl)

	response, requestErr := buildAndExecuteHttpRequest(requestWrapper, targetUrl)
	if requestErr != nil {
		fmt.Printf("RestAdapter: Error occured when making request. ", requestErr)
		return nil, requestErr
	} else if response.StatusCode == 404 {
		err := errors.New("Not found")
		return nil, err
	} else {
		defer response.Body.Close()

		// getting data from response
		data, ioErr := ioutil.ReadAll(response.Body)
		if ioErr != nil {
			fmt.Printf("RestAdapter: Error occured when reading response. ", ioErr)
			return nil, ioErr
		}

		// getting object out of the response
		object, objectParseErr := getJSONObjectFromResponse(data)
		if objectParseErr != nil {
			// if there is an error while getting object, try getting it as an array
			return nil, objectParseErr

		} else {
			return object, nil
		}
	}
	return nil, nil
}

func buildAndExecuteHttpRequest(requestWrapper message.RequestWrapper, url string) (resp *http.Response, err error) {
	client := &http.Client{}
	body, _ := json.Marshal(requestWrapper.Message.Body)
	request, _ := http.NewRequest(requestWrapper.Message.Command, url, bytes.NewBuffer(body))
	//	request, _ := http.NewRequest(requestWrapper.Message.Command, url, nil)
	request.Header.Set("X-Parse-Application-Id", parseAppId)
	request.Header.Set("X-Parse-REST-API-Key", parseApiKey)
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
