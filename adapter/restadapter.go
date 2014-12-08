package adapter

import (
	"github.com/eluleci/lightning/message"
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"errors"
)

const serverRoot = "http://api.maidan.co"

var arrayIdentifier string

type RestAdapter struct {

}

func (adapter *RestAdapter) ExecuteRequest(rw message.RequestWrapper) (map[string]interface{}, []map[string]interface{}, error) {

	arrayIdentifier = "data"

	var targetUrl string
	object := make(map[string]interface{})

	// TODO use custom converter for different end points
	if rw.Res == "/Discussion" {
		targetUrl = serverRoot+"/discussion/all?p=1&s=10&community=53d21f8f1d41c8127a000001&expand=*&filter=*"

	} else if rw.Res == "/Discussion/53d4b29962c77435d2000003" {
		targetUrl = serverRoot+"/discussion/id/53d4b29962c77435d2000003---"
	}

	response, requestErr := http.Get(targetUrl)
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

			if len(arrayIdentifier) > 0 {
				// if there is an arrayIdentifier in configuration, check the list field exists in object or not

				if objectInArrayIdentifierField, exists := object[arrayIdentifier]; exists {
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

func getJSONObjectFromResponse(inputData []byte) (object map[string]interface{}, err error) {

	err = json.Unmarshal(inputData, &object)
	return
}

func getJSONArrayFromResponse(inputData []byte) (objects []map[string]interface{}, err error) {

	err = json.Unmarshal(inputData, &objects)
	return
}
