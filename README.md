# ThunderDock
**ThunderDock** is a proxy server that turns an existing **REST API** into a **Real Time Web Sokcet Server** without writing any lines of code.

## How it works?
**ThunderDock** works with a simple messaging protocol on **Web Socket** connection. The message structure is similar to HTTP request structure. When client sends a data to the **ThunderDock** server, the server converts this message to an HTTP request and executes on the REST server. According to the response from the REST server, **ThunderDock** returns the response to the client and notifies other clients if there is anything to notify. So **ThunderDock** simply communicates between the clients and the server by forwarding the requests and publishing the result to all clients.

## Getting Started
**ThunderDock** server starts with initial configuration based on your REST API. It gets the configuration from a json file that is named "config.json" which contains the fields below:

**serverURL**(string):The root end-point of the REST API.

**objectIdentifier**(string): The key that defines the **id** of the items in the existing API. (ex: id, _id, objectId, itemId)

**collectionIdentifier**(string): The key that contains the array of the objects when requesting collections. If the root is already the array of items, you should leave this empty.

**persistItemInMemory**(boolean): If set **true**, **ThunderDock** server keeps the data of the item after the item is fetched for the first time. (caching)

**persistListInMemory**(boolean): If set **true**, **ThunderDock** server keeps the data of the item list after the list is fetched for the first time. (caching)

**cleanupOnSubscriptionsOver**(boolean): Destroys the object instance that is kept in the server if there is no more **Web Sokcet** connection is watching the object data.

An example 	**config.json** would be like this. (Real configuration for [Parse REST API](https://parse.com/docs/rest))

```
{
  "serverURL": "https://api.parse.com/1/classes",
  "objectIdentifier": "objectId",
  "collectionIdentifier": "results",
  "persistItemInMemory": false,
  "persistListInMemory": false,
  "cleanupOnSubscriptionsOver": true
}
```

**WARNING:** If the REST API needs a token(like Authentication token) from client for getting some item or list data, the values **persistItemInMemory** and **persistListInMemory** should be false. Because the data that is taken with the access token of a client can be accessed by another client who may not have a token. There will be a way provided in next versions for this problem.

# WS Server API Documentation

## Message Structures

**REQUEST MESSAGE STRUCTURE**

```
{
	rid: 123123123,         // identifier for the request, will be used for replying to the request
	res: 'RES',				// resource path of the object
	cmd: 'COMMAND',         // GET-POST-PUT-DELETE-HEAD
	headers: {              // optional
	},
	parameters: {           // optional
	},
    body: {                 // optional
    }
}
```

**RESPONSE MESSAGE STRUCTURE**

```
{
	rid: 123123123,         // identifier that is sent with the request message
	status: STATUS_CODE,    // status code (HTTP status codes)
	res: 'RES',             // sent as subscription id
    body: {                 // response body
    },
    error: {                // error object that contains detailed information about the error if error happens
    }
}
```

## Request Message

The methods that are used with **ThunderDock** are exactly same of HTTP Methods. The methods that are supported are GET, POST, PUT, DELETE and HEAD. The structure of the request messages are designed for creating an HTTP request from them. The examples below will explain how to create message for each method.

#### POST (Creation)

**REQUEST**

```
{
    "rid": 81775406,
    "cmd": "post",
    "res": "/Comment",
    "body": {
        "content": "new comment",
        "likes": 0
    }
}
```


**RESPONSE**

After creation, response and the body will contain the **res** of the created object which is inserted by the **ThunderDock** server for managing subscriptions.

```
{
    "rid": 81775406,
    "res": "/Comment/olIfG8duiP",
    "body": {
        "::res": "/Comment/olIfG8duiP",
        "createdAt": "2014-12-21T21:31:49.998Z",
        "objectId": "olIfG8duiP"
    },
    "status": 200
}
```

#### PUT (Update)

Request for update should only contain the changed fields of the object.

**REQUEST**

```
{
    "rid": 90244175,
    "cmd": "put",
    "res": "/Comment/xIDiatcdJk",
    "body": {
        "likes": 1
    }
}
```

**RESPONSE**

```
{
    "rid": 90244175,
    "res": "/Comment/xIDiatcdJk",
    "body": {
        "updatedAt": "2014-12-21T21:35:29.124Z"
    },
    "status": 200
}
```

#### GET

**REQUEST**

```
{
    "rid": 62599309,
    "cmd": "get",
    "res": "/Comment/xIDiatcdJk"
}
```

**RESPONSE**

```
{
    "rid": 62599309,
    "status": 200,
    "res": "/Comment/xIDiatcdJk",
    "body": {
        "::res": "/Comment/xIDiatcdJk",
        "content": "new comment",
        "createdAt": "2014-12-18T23:07:34.833Z",
        "likes": 23,
        "objectId": "xIDiatcdJk",
        "updatedAt": "2014-12-21T20:03:41.563Z"
    }
}
```

**Getting List Data**

If requested data is a list of items, then the body will contain a field **"::list"** which contains the array of items. 

**NOTE:** Connection will be subscribed to the **list** and to the all **individual items** in the list as well.

**REQUEST**

```
{
    "rid": 24820498,
    "cmd": "get",
    "res": "/Comment"
}
```

**RESPONSE**

```
{
    "rid": 24820498,
    "status": 200,
    "res": "/Comment",
    "body": {
        "::list": [
            {
                "::res": "/Comment/xIDiatcdJk",
                "content": "comment content",
                "createdAt": "2014-12-18T23:07:34.833Z",
                "likes": 23,
                "objectId": "xIDiatcdJk",
                "updatedAt": "2014-12-21T20:03:41.563Z"
            },
            {
                "::res": "/Comment/RSU3edhM2Q",
                "content": "comment content",
                "createdAt": "2014-12-21T14:51:22.913Z",
                "likes": 0,
                "objectId": "RSU3edhM2Q",
                "updatedAt": "2014-12-21T15:20:12.945Z"
            }
        ]
    }
}
```

#### DELETE

**REQUEST**

```
{
    "rid": 26195461,
    "cmd": "delete",
    "res": "/Comment/olIfG8duiP"
}
```

**RESPONSE***

```
{
    "rid": 26195461,
    "res": "/Comment/olIfG8duiP",
    "status": 200
}
```

## Push Message

**Web Socket** connection will receive a push message about the object that it is watching.

#### POST

If a connection is watching changes of a list, it will receive a push message when a new item created in that list.

The example below shows a push message when a new **Comment** object is created under the resource **/Comment**

```
{
    "res": "/Comment",
    "cmd": "post",
    "body": {
        "::res": "/Comment/cgX5fjrpAu",
        "content": "new comment",
        "createdAt": "2014-12-21T21:42:12.302Z",
        "likes": 0
    }
}
```

#### PUT

If connection is watching an object, it will receive a push message when that object is changed. The body of the object will only contain the changed fields of the object.

```
{
    "res": "/Comment/cgX5fjrpAu",
    "cmd": "put",
    "body": {
        "likes": 23,
        "updatedAt": "2014-12-21T21:46:06.818Z"
    }
}
```

#### DELETE
The connection will receive a push message about a deletion in two cases:

* when it's watching an object and the object is deleted

```
{
    "res": "/Comment/xIDiatcdJk",
    "cmd": "delete"
}
```
* when it's wathcing a list of objects and an item in the list is deleted

```
{
    "res": "/Comment",
    "cmd": "delete",
    "body": {
        "::res": "/Comment/xIDiatcdJk"
    }
}
```

**NOTE:** If the connection is watching both item and the list, it will receive both of the messages.

## Adding Query Parameters to Requests
It is possible to set parameters when getting an object list but **subscriptions** on the list that is retrieved with parameters is not supported yet. However the connection will still be subscribed to the **individual objects** in the list.

This is an example of getting **Comment** list ordered by likes (Parse API Example: **https://api.parse.com/1/classes/Comment?order=likes**)

**REQUEST**

```
{
    "rid": 17535539,
    "cmd": "get",
    "res": "/Comment",
    "parameters": {
        "order": ["likes"]
    }
}
```
**RESPONSE**

```
{
    "rid": 52081573,
    "res": "/Comment",
    "body": {
        "::list": [
            {
                "::res": "/Comment/cgX5fjrpAu",
                "content": "some comment",
                "createdAt": "2014-12-21T21:42:12.302Z",
                "likes": 23,
                "objectId": "cgX5fjrpAu",
                "updatedAt": "2014-12-21T21:46:06.818Z"
            },
            {
                "::res": "/Comment/olIfG8duiP",
                "content": "some other comment",
                "createdAt": "2014-12-21T21:31:49.998Z",
                "likes": 13,
                "objectId": "olIfG8duiP",
                "updatedAt": "2014-12-21T22:25:05.233Z"
            }
        ]
    },
    "status": 200
}
```

## Adding Headers to Requests
There are two ways of adding headers to requests.

###Setting headers for all requests
Clients can set headers by sending a message with a special command **::setHeaders** and the header map in the body. After setting, the headers will be inserted to all the requests starting from that moment.

**REQUEST**
	
```
{
    "rid": 90866495,
    "cmd": "::setHeaders",
    "body": {
        "X-Parse-Application-Id": [
            "oxjtnRnmGUKyM9SFd1szSKzO9wKHGlgC6WgyRpq8"
        ],
        "X-Parse-REST-API-Key": [
            "qJcOosDh8pywHdWKkVryWPoQFT1JMyoZYjMvnUul"
        ]
    }
}
```

**RESPONSE**

```
{
    "rid": 90866495,
    "status": 200
}
```

### Setting headers in individual requests
It is also possible to add the headers for each request by adding the headers to the request message. Response will be same as examples above.

**REQUEST**

```
{
    "rid": 10563647,
    "cmd": "get",
    "res": "/Comment/olIfG8duiP",
    "headers": {
        "X-Parse-Application-Id": [
            "oxjtnRnmGUKyM9SFd1szSKzO9wKHGlgC6WgyRpq8"
        ],
        "X-Parse-REST-API-Key": [
            "qJcOosDh8pywHdWKkVryWPoQFT1JMyoZYjMvnUul"
        ]
    }
}
```
**NOTE:** If the headers are set both generally and in a specific request, they will be combined while sending request to the REST API.

## Subscribing and Unsubscribing from resources
Conections will be subscribed to the resources automatically when they send any message to that resource. But it is also possible to subscribe and unsubscribe with special commands.

**REQUEST** (subscripton)

```
{
    "rid": 27207119,
    "cmd": "::subscribe",
    "res": "/Comment/olIfG8duiP"
}
```

**REQUEST** (unsubscription)

```
{
    "rid": 27207119,
    "cmd": "::unsubscribe",
    "res": "/Comment/olIfG8duiP"
}
```

**RESPONSE** (same for both)

```
{
    "rid": 27207119,
    "res": "/Comment/olIfG8duiP",
    "status": 200
}
```

