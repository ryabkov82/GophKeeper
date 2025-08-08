package connection

import (
	"context"
	"crypto/tls"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type fakeConn struct {
	stateFunc func() connectivity.State
	closeFunc func() error
}

func (f *fakeConn) GetState() connectivity.State {
	if f.stateFunc != nil {
		return f.stateFunc()
	}
	// По умолчанию возвращаем Ready
	return connectivity.Ready
}

func (f *fakeConn) Close() error {
	if f.closeFunc != nil {
		return f.closeFunc()
	}
	return nil
}

func (f *fakeConn) Connect() {}

func (f *fakeConn) WaitForStateChange(ctx context.Context, state connectivity.State) bool {
	return false
}

// grpc.ClientConnInterface методы-заглушки

func (f *fakeConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	return nil
}

func (f *fakeConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}

type fakeTLSConn struct{}

func (f *fakeTLSConn) ConnectionState() tls.ConnectionState {
	return tls.ConnectionState{} // можно заполнить нужные поля
}

func (f *fakeTLSConn) Close() error {
	return nil
}

func TestConnect_Success(t *testing.T) {
	logger := zap.NewNop()
	cfg := &Config{
		ServerAddress:  "localhost:1234",
		UseTLS:         false,
		ConnectTimeout: time.Second,
	}

	mgr := New(cfg, logger)

	// Мокаем dialFunc
	mgr.dialFunc = func(target string, opts ...grpc.DialOption) (GrpcConn, error) {
		return &fakeConn{}, nil
	}

	ctx := context.Background()
	conn, err := mgr.Connect(ctx)
	require.NoError(t, err)
	require.NotNil(t, conn)
}

func TestConnect_DialError(t *testing.T) {
	logger := zap.NewNop()
	cfg := &Config{
		ServerAddress:  "localhost:1234",
		UseTLS:         false,
		ConnectTimeout: time.Second,
	}

	mgr := New(cfg, logger)

	// Мокаем dialFunc с ошибкой
	mgr.dialFunc = func(target string, opts ...grpc.DialOption) (GrpcConn, error) {
		return nil, errors.New("dial error")
	}

	ctx := context.Background()
	conn, err := mgr.Connect(ctx)
	require.Error(t, err)
	require.Nil(t, conn)
}

func TestTestTLSConnection_Success(t *testing.T) {
	logger := zap.NewNop()
	cfg := &Config{
		ServerAddress: "localhost:443",
		UseTLS:        true,
	}

	mgr := New(cfg, logger)

	// Мокаем tlsDialFunc, возвращаем успешное соединение
	mgr.tlsDialFunc = func(network, addr string, config *tls.Config) (tlsConn, error) {
		return &fakeTLSConn{}, nil
	}

	err := mgr.testTLSConnection(&tls.Config{})
	require.NoError(t, err)
}

func TestTestTLSConnection_Error(t *testing.T) {
	logger := zap.NewNop()
	cfg := &Config{
		ServerAddress: "invalid:address",
		UseTLS:        true,
	}

	mgr := New(cfg, logger)

	// Мокаем tlsDialFunc, возвращаем ошибку
	mgr.tlsDialFunc = func(network, addr string, config *tls.Config) (tlsConn, error) {
		return nil, errors.New("tls dial error")
	}

	err := mgr.testTLSConnection(&tls.Config{})
	require.Error(t, err)
}

func TestClose_WithOpenConnection(t *testing.T) {
	logger := zap.NewNop()
	cfg := &Config{
		ServerAddress: "localhost:1234",
	}

	mgr := New(cfg, logger)

	closed := false
	mgr.conn = &fakeConn{
		closeFunc: func() error {
			closed = true
			return nil
		},
		stateFunc: func() connectivity.State {
			return connectivity.Ready
		},
	}

	err := mgr.Close()
	require.NoError(t, err)
	require.True(t, closed, "expected Close() to be called on connection")
}

func TestClose_NoConnection(t *testing.T) {
	logger := zap.NewNop()
	cfg := &Config{
		ServerAddress: "localhost:1234",
	}

	mgr := New(cfg, logger)

	err := mgr.Close()
	require.NoError(t, err) // ничего не должно происходить, ошибки нет
}

func TestIsReady_TrueAndFalse(t *testing.T) {
	logger := zap.NewNop()
	cfg := &Config{
		ServerAddress: "localhost:1234",
	}

	mgr := New(cfg, logger)

	// Случай, когда conn == nil
	require.False(t, mgr.IsReady())

	// Подменяем conn с состоянием Ready
	mgr.conn = &fakeConn{
		stateFunc: func() connectivity.State {
			return connectivity.Ready
		},
	}
	require.True(t, mgr.IsReady())

	// Подменяем conn с состоянием не Ready
	mgr.conn = &fakeConn{
		stateFunc: func() connectivity.State {
			return connectivity.Connecting
		},
	}
	require.False(t, mgr.IsReady())
}

func TestConn(t *testing.T) {
	logger := zap.NewNop()
	cfg := &Config{
		ServerAddress: "localhost:1234",
	}

	mgr := New(cfg, logger)
	require.Nil(t, mgr.Conn())

	dummyConn := &grpc.ClientConn{}
	mgr.conn = dummyConn
	require.Equal(t, dummyConn, mgr.Conn())
}

func TestManager_Close(t *testing.T) {
	logger := zap.NewNop()
	cfg := &Config{ServerAddress: "localhost:1234"}

	mgr := New(cfg, logger)

	closed := false
	mgr.conn = &fakeConn{
		closeFunc: func() error {
			closed = true
			return nil
		},
	}

	err := mgr.Close()
	require.NoError(t, err)
	require.True(t, closed, "expected Close() to be called on connection")
}

func TestManager_Close_Error(t *testing.T) {
	logger := zap.NewNop()
	cfg := &Config{ServerAddress: "localhost:1234"}

	mgr := New(cfg, logger)

	mgr.conn = &fakeConn{
		closeFunc: func() error {
			return errors.New("close failed")
		},
	}

	err := mgr.Close()
	require.Error(t, err)
	require.EqualError(t, err, "close failed")
}

func TestManager_Close_NilConn(t *testing.T) {
	logger := zap.NewNop()
	cfg := &Config{ServerAddress: "localhost:1234"}

	mgr := New(cfg, logger)

	// conn == nil, Close должен вернуть nil и ничего не делать
	err := mgr.Close()
	require.NoError(t, err)
}
