// nolint:testifylint,gosec
package gocket_test

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Falokut/go-kit/gocket"
	"github.com/coder/websocket"
	"github.com/stretchr/testify/require"
)

func TestClient_OnDisconnect(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	srvHandler := NewCallHandler()
	srv := gocket.NewServer(gocket.WithServerAcceptOptions(
		&websocket.AcceptOptions{
			InsecureSkipVerify: true,
		},
	))
	srv.OnDisconnect(srvHandler.OnDisconnect)

	httpSrv := http.Server{
		Handler: srv,
		Addr:    "localhost:7878",
	}
	go func() {
		_ = httpSrv.ListenAndServe()
	}()
	time.Sleep(500 * time.Millisecond)

	cliHandler := NewCallHandler()
	cli := gocket.NewClient()
	err := cli.OnDisconnect(cliHandler.OnDisconnect).
		OnConnect(cliHandler.OnConnect).
		Dial(context.Background(), "ws://localhost:7878")
	require.NoError(err)
	time.Sleep(500 * time.Millisecond)

	err = httpSrv.Shutdown(context.Background())
	require.NoError(err)
	time.Sleep(500 * time.Millisecond)

	require.EqualValues(0, cliHandler.onDisconnectCount.Load())
	require.EqualValues(1, cliHandler.onConnectCount.Load())

	srv.Shutdown()
	time.Sleep(500 * time.Millisecond)

	require.EqualValues(1, cliHandler.onDisconnectCount.Load())
	require.EqualValues(1, srvHandler.onDisconnectCount.Load())
}

func TestClientHeavyConcurrency(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	_, srv, testServer := serve(t, nil, nil)

	time.Sleep(300 * time.Millisecond)

	srvHandler := NewCallHandler()
	srv.OnConnect(func(conn *gocket.Conn) {
		srvHandler.OnConnect(conn)
		srv.Rooms().Join(conn, "connections")
		err := conn.Emit(context.Background(), "hello", nil)
		require.NoError(err)

		conn.Data().Add("key", conn.Id())
	}).On("closeMe", gocket.HandlerFunc(func(ctx context.Context, conn *gocket.Conn, msg gocket.Message) []byte {
		connId := conn.Data().Get("key")
		require.Equal(conn.Id(), connId)

		err := conn.Close()
		require.NoError(err)
		return nil
	})).OnDisconnect(srvHandler.OnDisconnect)

	wg := sync.WaitGroup{}
	for range 100 {
		wg.Add(1)
		go func() {
			cli := gocket.NewClient(
				gocket.WithClientDialOptions(
					&websocket.DialOptions{
						HTTPClient: testServer.Client(),
					},
				),
			)

			cli.On("hello", gocket.HandlerFunc(func(ctx context.Context, conn *gocket.Conn, msg gocket.Message) []byte {
				err := conn.Emit(context.Background(), "closeMe", nil)
				require.NoError(err)
				return nil
			}))
			cli.OnDisconnect(func(conn *gocket.Conn, err error) {
				emitError := cli.Emit(context.Background(), "closeMe", nil)
				require.Error(emitError)
				wg.Done()
			})

			err := cli.Dial(context.Background(), strings.ReplaceAll(testServer.URL, "http", "ws"))
			require.NoError(err)
		}()
		time.Sleep(100 * time.Millisecond)
	}

	waitAllDisconnections := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitAllDisconnections)
	}()

	select {
	case <-waitAllDisconnections:
		time.Sleep(1 * time.Second)
		require.Len(srv.Rooms().AllConns(), 1) // first client in serve
		require.Len(srv.Rooms().Rooms(), 0)
		require.EqualValues(100, srvHandler.onConnectCount.Load())
		require.EqualValues(100, srvHandler.onDisconnectCount.Load())
		require.EqualValues(0, srvHandler.onErrorCount.Load())
	case <-time.After(5 * time.Second):
		require.Fail("wait all disconnections timeout")
	}
}
