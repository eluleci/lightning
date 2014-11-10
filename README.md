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

## Create Profile
While joining to the system for the first time

**REQUEST**

```
{
	rid: 123123123,
	cmd: 'create',
    body: {
    	id:	'jsdjNBKJSBDalsdhbasdabnslduaDlsajdn',
    	name: 'Emrullah',
        surname: 'Lüleci',
        avatar: 'https://graph.facebook.com/eluleci/picture?type=large'
    }
}
```

## Updating Data


**REQUEST**

```
{
	tx: 123123123,
	cmd: 'update',
	data: {
        image: ''
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