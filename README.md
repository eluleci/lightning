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


# API DOCUMENTATION

## Create Object

**REQUEST**

```
{
	rid: 123123123,
	cmd: 'post',
	res: 'Profile',
    body: {
    	className: 'Profile',       // domain of the object type
        key: 'value',
    }
}
```

**RESPONSE**

```
{
	rid: 123123123,
	cmd: 'post',
	status: 200,                    // status code
    body: {
    	id: 'ID',                   // id of the new object
    }
}
```

## Updating Data


**REQUEST**

```
{
	rid: 123123123,
	cmd: 'post',
	body: {
	    id: 'ID',                   // id of the object to update
    	className: 'Profile',       // domain of the object type
        key: 'value',
	}
}
```

## RESPONSE TYPES

**SUCCESSFUL RESPONSE**

```
{
	rid: 123123123,
	code: 201
}
```

**FAIL RESPONSE**

```
{
	rid: 123123123,
	code: 400
}
```