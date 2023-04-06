package emitter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cuizuoli/socket.io-redis-emitter-go/enum"
	"github.com/go-redis/redis/v8"
	"github.com/vmihailenco/msgpack/v5"
)

type BroadcastOptions struct {
	nsp              string
	broadcastChannel string
	requestChannel   string
}

type BroadcastFlags struct {
	volatile bool
	compress bool
}

var ReservedEvents = map[string]interface{}{
	"connect":        nil,
	"connect_error":  nil,
	"disconnect":     nil,
	"disconnecting":  nil,
	"newListener":    nil,
	"removeListener": nil,
}

type BroadcastOperator struct {
	redisClient *redis.Client
	opts        BroadcastOptions
	rooms       map[string]interface{}
	exceptRooms map[string]interface{}
	flags       BroadcastFlags
}

func NewBroadcastOperator(redisClient *redis.Client, opts BroadcastOptions) *BroadcastOperator {
	return &BroadcastOperator{
		redisClient: redisClient,
		opts:        opts,
		rooms:       make(map[string]interface{}),
		exceptRooms: make(map[string]interface{}),
		flags:       BroadcastFlags{},
	}
}

/**
 * Targets a room when emitting.
 *
 * @param room
 * @return a new BroadcastOperator instance
 * @public
 */
func (b *BroadcastOperator) To(rooms ...string) *BroadcastOperator {
	for _, room := range rooms {
		b.rooms[room] = nil
	}
	return b
}

/**
 * Targets a room when emitting.
 *
 * @param room
 * @return a new BroadcastOperator instance
 * @public
 */
func (b *BroadcastOperator) In(rooms ...string) *BroadcastOperator {
	return b.To(rooms...)
}

/**
 * Excludes a room when emitting.
 *
 * @param room
 * @return a new BroadcastOperator instance
 * @public
 */
func (b *BroadcastOperator) Except(rooms ...string) *BroadcastOperator {
	for _, room := range rooms {
		b.exceptRooms[room] = nil
	}
	return b
}

/**
 * Sets the compress flag.
 *
 * @param compress - if `true`, compresses the sending data
 * @return a new BroadcastOperator instance
 * @public
 */
func (b *BroadcastOperator) Compress(compress bool) *BroadcastOperator {
	b.flags.compress = compress
	return b
}

/**
 * Sets a modifier for a subsequent event emission that the event data may be lost if the client is not ready to
 * receive messages (because of network slowness or other issues, or because they’re connected through long polling
 * and is in the middle of a request-response cycle).
 *
 * @return a new BroadcastOperator instance
 * @public
 */
func (b *BroadcastOperator) Volatile() *BroadcastOperator {
	b.flags.volatile = true
	return b
}

/**
 * Emits to all clients.
 *
 * @return Always true
 * @public
 */
func (b *BroadcastOperator) Emit(event string, args ...interface{}) (bool, error) {
	if _, ok := ReservedEvents[event]; ok {
		return false, errors.New(fmt.Sprintf("%s is a reserved event name", event))
	}

	// set up packet object
	data := append([]interface{}{event}, args...)
	packet := map[string]interface{}{
		"type": enum.PacketTypeEvent,
		"data": data,
		"nsp":  b.opts.nsp,
	}

	rooms := mapToArray(b.rooms)
	opts := map[string]interface{}{
		"rooms":  rooms,
		"flags":  b.flags,
		"except": mapToArray(b.exceptRooms),
	}

	msg, err := msgpack.Marshal([]interface{}{uid, packet, opts})
	if err != nil {
		return false, err
	}

	channel := b.opts.broadcastChannel
	if len(rooms) == 1 {
		channel = fmt.Sprintf("%s%s#", channel, rooms[0])
	}

	ctx := context.Background()
	err = b.redisClient.Publish(ctx, channel, msg).Err()
	if err != nil {
		return false, err
	}

	return true, nil
}

/**
 * Makes the matching socket instances join the specified rooms
 *
 * @param rooms
 * @public
 */
func (b *BroadcastOperator) SocketsJoin(rooms ...string) error {
	request, err := json.Marshal(map[string]interface{}{
		"type": enum.RequestTypeRemoteJoin,
		"opts": map[string][]string{
			"rooms":  mapToArray(b.rooms),
			"except": mapToArray(b.exceptRooms),
		},
		"rooms": rooms,
	})
	if err != nil {
		return err
	}
	ctx := context.Background()
	return b.redisClient.Publish(ctx, b.opts.requestChannel, request).Err()
}

/**
 * Makes the matching socket instances leave the specified rooms
 *
 * @param rooms
 * @public
 */
func (b *BroadcastOperator) SocketsLeave(rooms ...string) error {
	request, err := json.Marshal(map[string]interface{}{
		"type": enum.RequestTypeRemoteLeave,
		"opts": map[string][]string{
			"rooms":  mapToArray(b.rooms),
			"except": mapToArray(b.exceptRooms),
		},
		"rooms": rooms,
	})
	if err != nil {
		return err
	}
	ctx := context.Background()
	return b.redisClient.Publish(ctx, b.opts.requestChannel, request).Err()
}

/**
 * Makes the matching socket instances disconnect
 *
 * @param close - whether to close the underlying connection
 * @public
 */
func (b *BroadcastOperator) DisconnectSockets(close bool) error {
	request, err := json.Marshal(map[string]interface{}{
		"type": enum.RequestTypeRemoteDisconnect,
		"opts": map[string][]string{
			"rooms":  mapToArray(b.rooms),
			"except": mapToArray(b.exceptRooms),
		},
		"close": close,
	})
	if err != nil {
		return err
	}
	ctx := context.Background()
	return b.redisClient.Publish(ctx, b.opts.requestChannel, request).Err()
}

/**
 * Send a packet to the Socket.IO servers in the cluster
 *
 * @param args - any number of serializable arguments
 */
func (b *BroadcastOperator) ServerSideEmit(args ...interface{}) error {
	request, err := json.Marshal(map[string]interface{}{
		"uid":  uid,
		"type": enum.RequestTypeServerSideEmit,
		"data": args,
	})
	if err != nil {
		return err
	}
	ctx := context.Background()
	return b.redisClient.Publish(ctx, b.opts.requestChannel, request).Err()
}

func mapToArray(m map[string]interface{}) []string {
	var rooms []string
	if len(m) > 0 {
		for room := range m {
			rooms = append(rooms, room)
		}
	}
	return rooms
}
