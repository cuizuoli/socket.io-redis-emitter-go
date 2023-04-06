package emitter

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"strings"
	"time"
)

const uid = "emitter"
const defaultKey = "socket.io"
const defaultNsp = "/"

type RedisOptions struct {
	Address     string
	Password    string
	ReadTimeout int
	PoolSize    int
}

type EmitterOptions struct {
	Redis RedisOptions
	Key   string
	Nsp   string
}

type Emitter struct {
	opts             EmitterOptions
	redisClient      *redis.Client
	broadcastOptions BroadcastOptions
}

func NewEmitter(opts EmitterOptions) *Emitter {
	redisClient := redis.NewClient(&redis.Options{
		Addr:        opts.Redis.Address,
		Password:    opts.Redis.Password,
		ReadTimeout: time.Duration(opts.Redis.ReadTimeout) * time.Second,
		PoolSize:    opts.Redis.PoolSize,
	})
	if opts.Key == "" {
		opts.Key = defaultKey
	}
	if opts.Nsp == "" {
		opts.Nsp = defaultNsp
	}
	broadcastOptions := BroadcastOptions{
		nsp:              opts.Nsp,
		broadcastChannel: fmt.Sprintf("%s#%s#", opts.Key, opts.Nsp),
		requestChannel:   fmt.Sprintf("%s-request#%s#", opts.Key, opts.Nsp),
	}
	return &Emitter{
		opts:             opts,
		redisClient:      redisClient,
		broadcastOptions: broadcastOptions,
	}
}

/**
 * Return a new emitter for the given namespace.
 *
 * @param nsp - namespace
 * @public
 */
func (e *Emitter) Of(nsp string) *Emitter {
	if !strings.HasPrefix(nsp, defaultNsp) {
		nsp = defaultNsp + nsp
	}
	e.opts.Nsp = nsp
	return NewEmitter(e.opts)
}

/**
 * Emits to all clients.
 *
 * @return Always true
 * @public
 */
func (e *Emitter) Emit(event string, args ...interface{}) (bool, error) {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions).Emit(event, args)
}

/**
 * Targets a room when emitting.
 *
 * @param room
 * @return BroadcastOperator
 * @public
 */
func (e *Emitter) To(rooms ...string) *BroadcastOperator {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions).To(rooms...)
}

/**
 * Targets a room when emitting.
 *
 * @param room
 * @return BroadcastOperator
 * @public
 */
func (e *Emitter) In(rooms ...string) *BroadcastOperator {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions).In(rooms...)
}

/**
 * Excludes a room when emitting.
 *
 * @param room
 * @return BroadcastOperator
 * @public
 */
func (e *Emitter) Except(rooms ...string) *BroadcastOperator {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions).Except(rooms...)
}

/**
 * Sets a modifier for a subsequent event emission that the event data may be lost if the client is not ready to
 * receive messages (because of network slowness or other issues, or because they’re connected through long polling
 * and is in the middle of a request-response cycle).
 *
 * @return BroadcastOperator
 * @public
 */
func (e *Emitter) Volatile() *BroadcastOperator {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions).Volatile()
}

/**
 * Sets the compress flag.
 *
 * @param compress - if `true`, compresses the sending data
 * @return BroadcastOperator
 * @public
 */
func (e *Emitter) Compress(compress bool) *BroadcastOperator {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions).Compress(compress)
}

/**
 * Makes the matching socket instances join the specified rooms
 *
 * @param rooms
 * @public
 */
func (e *Emitter) SocketsJoin(rooms ...string) error {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions).SocketsJoin(rooms...)
}

/**
 * Makes the matching socket instances leave the specified rooms
 *
 * @param rooms
 * @public
 */
func (e *Emitter) SocketsLeave(rooms ...string) error {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions).SocketsLeave(rooms...)
}

/**
 * Makes the matching socket instances disconnect
 *
 * @param close - whether to close the underlying connection
 * @public
 */
func (e *Emitter) DisconnectSockets(close bool) error {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions).DisconnectSockets(close)
}

/**
 * Send a packet to the Socket.IO servers in the cluster
 *
 * @param args - any number of serializable arguments
 */
func (e *Emitter) ServerSideEmit(args ...interface{}) error {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions).ServerSideEmit(args...)
}
