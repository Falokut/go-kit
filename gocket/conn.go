package gocket

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Falokut/go-kit/ds"
	"github.com/Falokut/go-kit/gocket/internal"
	"github.com/coder/websocket"
)

type Conn struct {
	id      string
	request *http.Request
	ws      *websocket.Conn
	data    *ds.SafeMap[string, any]
	acks    *internal.Acks
}

func newConn(
	id string,
	request *http.Request,
	ws *websocket.Conn,
) *Conn {
	return &Conn{
		id:      id,
		request: request,
		ws:      ws,
		data:    ds.NewSafeMap[string, any](),
		acks:    internal.NewAcks(),
	}
}

func (c *Conn) Id() string {
	return c.id
}

func (c *Conn) HttpRequest() *http.Request {
	return c.request
}

func (c *Conn) Data() *ds.SafeMap[string, any] {
	return c.data
}

func (c *Conn) Emit(ctx context.Context, event string, data []byte) error {
	message := Message{
		Event: event,
		Data:  data,
		AckId: 0,
	}
	return c.emit(ctx, message)
}

func (c *Conn) EmitWithAck(ctx context.Context, event string, data []byte) ([]byte, error) {
	ack := c.acks.NextAck()
	defer c.acks.DeleteAck(ack.Id())

	message := Message{
		Event: event,
		Data:  data,
		AckId: ack.Id(),
	}
	err := c.emit(ctx, message)
	if err != nil {
		return nil, err
	}

	response, err := ack.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait ack: %w", err)
	}

	return response, nil
}

func (c *Conn) Ping(ctx context.Context) error {
	return c.ws.Ping(ctx)
}

func (c *Conn) Close() error {
	return c.ws.Close(websocket.StatusNormalClosure, "")
}

func (c *Conn) emit(ctx context.Context, msg Message) error {
	buff := getBuffer()
	defer putInBuffer(buff)

	msg.Encode(buff)

	err := c.ws.Write(ctx, websocket.MessageText, buff.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	return nil
}

func (c *Conn) notifyAck(ackId uint64, data []byte) {
	c.acks.NotifyAck(ackId, data)
}
