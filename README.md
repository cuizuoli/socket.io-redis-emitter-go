socket.io-redis-emitter-go
=====================

A Golang implementation of [socket.io-redis-emitter](https://github.com/socketio/socket.io-redis-emitter)

This project uses redis.
Make sure your environment has redis.

Install and development
--------------------

To install in your golang project.

```sh
$ go get github.com/cuizuoli/socket.io-redis-emitter-go
```

Usage
---------------------

Example:

```go
e := emitter.NewEmitter(emitter.EmitterOptions{
    Redis: emitter.RedisOptions{
        Address:     "localhost:6379",
        Password:    "123456",
        ReadTimeout: 30,
        PoolSize:    8,
    },
})
e.Emit("message", "Hi!")
```
