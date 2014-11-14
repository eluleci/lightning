# Chat Example

This application shows how to use use the
[websocket](https://github.com/gorilla/websocket) package and
[jQuery](http://jquery.com) to implement a simple web chat application.

## Running the example

The example requires a working Go development environment. The [Getting
Started](http://golang.org/doc/install) page describes how to install the
development environment.

Once you have Go up and running, you can download, build and run the example
using the following commands.

    $ go get github.com/gorilla/websocket
    $ cd `go list -f '{{.Dir}}' github.com/gorilla/websocket/examples/chat`
    $ go run *.go


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