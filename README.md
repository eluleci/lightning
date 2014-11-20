# API DOCUMENTATION

## Messages

**REQUEST MESSAGE STRUCTURE**

```
{
	rid: 123123123,         // identifier for the request, will be used for replying to the request
	cmd: 'COMMAND',         // GET-POST-PUT-DELETE
	res: 'RES',
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
	headers: {              // response headers
	},
    body: {                 // response body
    },
    error: {                // error object that contains detailed information about the error, if there is an error
    }
}
```

## Creating Object

**REQUEST**

When creating a new object, if there is no specific resource domain, the **res** must be the **className**

```
{
	rid: 123123123,
	cmd: 'post',
	res: '/Profile',
    body: {
    	className: 'Profile',
        key: 'value',
    }
}
```

If the object will be created under a specific domain, **res** must be that domain.

```
{
	rid: 123123123,
	cmd: 'post',
	res: '/Profile/1/Address',
    body: {
    	className: 'Address',
        key: 'value',
    }
}
```

**RESPONSE**

After creation, response will contain the **res** of the created object. Also body will contain the id and res.

```
{
	rid: 123123123,
	status: 200,                        // status code
	res: '/Profile/1/Address/ID',
    body: {
	    res: '/Profile/1/Address/ID',   // full res of the object
    	id: 'ID',                       // id of the new object
    }
}
```

## Updating Data

Update object message must contain the **res** of the object in message and also in the object body. The
main differentiation between the creation and update is that the request body contains the res or not. If there is no
res inside the body, this will be accepted as a **create message** for a new object under the given **res**.

**REQUEST**

```
{
	rid: 123123123,
	cmd: 'post',
	res: '/Profile/1/Address/ID',
	body: {
	    res: '/Profile/1/Address/ID',
        key: 'value',
	}
}
```

**RESPONSE**

After successful update, response message will contain only **rid**,**status**, and **res**. So main difference
between the create or update responses is that there is a body in the response or not.

```
{
	rid: 123123123,
	status: 200,                    // status code
	res: '/Profile/1/Address/ID'
}
```

## Getting Data

Get message must contain a valid **res** for the object or the list. The body of the response will contain the data of
the object body.

**REQUEST**

```
{
	rid: 123123123,
	cmd: 'get',
	res: '/Profile/1'
}
```

**RESPONSE**

```
{
	rid: 123123123,
	status: 200,
	res: '/Profile/1',
    body: {
	    res: '/Profile/1',
    	id: 'ID',
    	className: 'Profile',
    	key: 'value'
    }
}
```

If requested data is a list of objects, then the body will contain a field **"loList"** which contains
the list of objects.

**REQUEST**

```
{
	rid: 123123123,
	cmd: 'get',
	res: '/Profile/1/Address'
}
```

**RESPONSE**

```
{
	rid: 123123123,
	status: 200,
	res: '/Profile/1/Address',
    body: {
    	loList: [
    	    {
	            res: '/Profile/1/Address/1',
                id: 'ID',
                className: 'Address',
                key: 'value'
            },
            {
	            res: '/Profile/1/Address/2',
                id: 'ID',
                className: 'Address',
                key: 'value'
            },
    	]
    }
}
```

## Getting Push Message

Connection will receive a push message about the object that they're watching.

### Creation Push

If connection is watching changes of a list, the connection will receive a push message when a new item created.
The **res** of the message is the **res** of the list and the data of the object will be in the body of the message.
Assume the **res** of the list is **/Profile** and a new item is created under this domain. The message would be;

**PUSH MESSAGE**

```
{
	cmd: 'post',
	res: '/Profile'                 // the resource that the object is created in
    body: {                         // data of the created item
	    res: '/Profile/1',
    	id: 'ID',
    	className: 'Profile',
    	key: 'value'
    }
}
```

### Update Push

If connection is watching an object, it will be notified when there is a change in the object. The **res** of the message
is the **res** of the object and the updated fields are in the body of the message. Only updated fields will be inside
the body.

**PUSH MESSAGE**

```
{
	cmd: 'post',
	res: '/Profile/1'               // the resource that the object is created in
    body: {
    	key: 'value'                // changed fields
    }
}
```

## Adding references

When an object needs to keep reference, the field just need to contain the **res** of that reference. For distinguishing
the reference from any other string, the reference is saved with the prefix **"loRef::"**. So the reference field will
contain a string like **"loRef::OBJECT_RES"**. As an example, a comment object with an **author** reference would be
like:

```
{
	id: OBJECT_ID,
	className: "Comment",
	text: "This is a comment text",
	likes: 13,
	author: "loRef::/Profile/1"
}
```


# SYSTEM STRUCTURE

## Actors

### Main
Is responsbile of

* creating web socket connections and storing them
* creating main hub

It watches one channel

* **seekHub(<-requestWrapper)**: Finds the requested hub object that is defined inside the message of requestWrapper. Then it gives the requestWrapper to the hub to handle.

### Connection
Is responsible of

* reading and writing messages to web socket connection
* keeping references to all hubs that the connection is subscribed to

It watches two channels;

* **send**: Waits for a message to send it to the connection buffer
* **implicitSubscription(<-hub)**: Connections are automatically subscribed to an object that is created inside the list that they are subscribed.
* **implicitUnsubscription(<-hub)**: Connections are automatically unsubscribed if the object has been deleted or it is no longer available for subscription.

### Hub
Is responsible of

* adding and removing subscriptions to itself
* keeping references to all connections that are subscribed to it
* sending broadcast messages to all connections in it
* notifying the **domain hub** if the object is just created. (When the hub object is first created, it needs to notify the hub object which handles the list of the objects for that reference.)

It watches three channels

* **subscribe(requestWrapper)**: Gets a request wrapper object and adds the **listener channel** to subscriptions. Then sends the message in the wrapper to ModelHolder.
* **unsubscribe(channel)**: Gets a channel and removes it from the subscription list.
* **broadcast(Message)**: Gets a message and sends it to all subscribed connections.

### ModelHolder
Is responsible of

* keeping the data of the object
* keeping the reference to all ModelHolders in the list if it is a domain ModelHolder
* handling all messages to itself and return response to the given channel
* requesting broadcast if there is a change that all connections need to receive

Wathces one channel

* **handle(requestWrapper)**: Gets a message, applies the changes and returns a response message. Also broadcasts a message to hub if needed.


