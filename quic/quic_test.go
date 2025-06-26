// nolint:gosec
package quic_test

import (
	"context"
	"crypto/tls"
	"io"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Falokut/go-kit/quic"
	"github.com/Falokut/go-kit/quic/client"
	quiclib "github.com/quic-go/quic-go"
	"github.com/stretchr/testify/suite"
)

func TestQUICSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(QUICSuite))
}

type QUICSuite struct {
	suite.Suite

	addr   string
	server *quic.Server
	cancel context.CancelFunc
}

func (s *QUICSuite) SetupSuite() {
	_, err := os.Stat("test_cert.pem")
	s.False(os.IsNotExist(err),
		"Missing test_cert.pem — run openssl to generate")

	testCert, err := os.ReadFile("test_cert.pem")
	s.Require().NoError(err)

	testKey, err := os.ReadFile("test_key.pem")
	s.Require().NoError(err)

	quicConf := &quiclib.Config{
		EnableDatagrams:    false,
		MaxIncomingStreams: 100,
	}

	cert, err := tls.X509KeyPair(testCert, testKey)
	s.Require().NoError(err)

	s.server = quic.NewServer(
		quic.WithTLSConfig(&tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"h3"},
		}),
		quic.WithQUICConfig(quicConf),
	)

	s.server.Upgrade(quic.StreamHandlerFunc(echoHandler))

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go func() {
		err := s.server.ListenAndServe(ctx, "localhost:0")
		s.NoError(err)
	}()

	for {
		if s.server.Listener() != nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	s.addr = s.server.Listener().Addr().String()
}

func (s *QUICSuite) TearDownSuite() {
	s.cancel()
	_ = s.server.Shutdown(context.Background())
}

func echoHandler(ctx context.Context, stream quic.Stream) error {
	defer stream.Close()
	body, err := io.ReadAll(stream)
	if err != nil {
		return err
	}
	_, err = stream.Write([]byte("echo: " + string(body)))
	return err
}

func (s *QUICSuite) TestQUICEcho() {
	c := client.NewClient()
	resp, err := c.Request(s.addr).Body([]byte("hello")).Do(context.Background())
	s.Require().NoError(err)
	s.Equal("echo: hello", string(resp.Body))
}

func (s *QUICSuite) TestMiddleware() {
	invoked := atomic.Bool{}

	loggingMW := func(next client.RoundTripper) client.RoundTripper {
		return client.RoundTripperFunc(func(ctx context.Context, req *quic.Request) (*quic.Response, error) {
			invoked.Store(true)
			return next.RoundTrip(ctx, req)
		})
	}

	c := client.NewClient(client.WithMiddlewares(loggingMW))
	resp, err := c.Request(s.addr).Body([]byte("ping")).Do(context.Background())
	s.Require().NoError(err)
	s.Equal("echo: ping", string(resp.Body))
	s.True(invoked.Load(), "middleware must be invoked")
}
