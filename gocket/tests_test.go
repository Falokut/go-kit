package gocket_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Falokut/go-kit/gocket"
	"github.com/coder/websocket"
)

type CallHandler struct {
	onConnectCount    atomic.Int32
	onDisconnectCount atomic.Int32
	onErrorCount      atomic.Int32
	events            map[string][][]byte
	lock              sync.Locker
}

func NewCallHandler() *CallHandler {
	return &CallHandler{
		events: make(map[string][][]byte),
		lock:   &sync.Mutex{},
	}
}

func (c *CallHandler) OnConnect(conn *gocket.Conn) {
	c.onConnectCount.Add(1)
}

func (c *CallHandler) OnDisconnect(conn *gocket.Conn, err error) {
	c.onDisconnectCount.Add(1)
}

func (c *CallHandler) OnError(conn *gocket.Conn, err error) {
	c.onErrorCount.Add(1)
}

// nolint:ireturn
func (c *CallHandler) Handle(response []byte) gocket.Handler {
	return gocket.HandlerFunc(func(ctx context.Context, conn *gocket.Conn, msg gocket.Message) []byte {
		c.lock.Lock()
		defer c.lock.Unlock()

		c.events[msg.Event] = append(c.events[msg.Event], msg.Data)
		return response
	})
}

func serve(
	t *testing.T,
	srvCfg func(srv *gocket.Server),
	cliCfg func(cli *gocket.Client),
) (*gocket.Client, *gocket.Server, *httptest.Server) {
	t.Helper()

	srv := gocket.NewServer(gocket.WithServerAcceptOptions(
		&websocket.AcceptOptions{
			InsecureSkipVerify: true,
		},
	))
	if srvCfg != nil {
		srvCfg(srv)
	}
	testServer := httptest.NewServer(srv)
	t.Cleanup(func() {
		testServer.Close()
		srv.Shutdown()
	})

	cli := gocket.NewClient(gocket.WithClientDialOptions(
		&websocket.DialOptions{
			HTTPClient: testServer.Client(),
		},
	))
	if cliCfg != nil {
		cliCfg(cli)
	}
	err := cli.Dial(context.Background(), strings.ReplaceAll(testServer.URL, "http", "ws"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cli.Close()
	})

	return cli, srv, testServer
}
