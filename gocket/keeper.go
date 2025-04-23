package gocket

import (
	"context"

	"github.com/Falokut/go-kit/gocket/internal"
)

type keeper struct {
	conn *Conn
	mux  *mux
}

func newKeeper(conn *Conn, mux *mux) *keeper {
	return &keeper{
		conn: conn,
		mux:  mux,
	}
}

func (k *keeper) Serve(ctx context.Context) {
	k.mux.handleConnect(k.conn)
	for {
		err := k.readAndHandleMessage(ctx)
		if err != nil {
			k.mux.handleDisconnect(k.conn, err)
			return
		}
	}
}

func (k *keeper) readAndHandleMessage(ctx context.Context) error {
	_, data, err := k.conn.ws.Read(ctx)
	if err != nil {
		return err
	}

	msg := Message{}
	err = msg.Unmarshal(data)
	if err != nil {
		k.mux.onError(k.conn, err)
		return nil
	}

	if internal.IsAckEvent(msg.Event) {
		k.conn.notifyAck(msg.AckId, msg.Data)
		return nil
	}

	go func() {
		response := k.mux.handle(ctx, k.conn, msg)
		if !msg.IsAckRequired() {
			return
		}

		message := Message{
			Event: internal.ToAckEvent(msg.Event),
			AckId: msg.AckId,
			Data:  response,
		}
		err := k.conn.emit(ctx, message)
		if err != nil {
			k.mux.onError(k.conn, err)
		}
	}()

	return nil
}
