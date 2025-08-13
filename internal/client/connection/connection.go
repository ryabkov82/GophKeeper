package connection

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"os"
	"sync"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/client/service/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Config содержит параметры конфигурации подключения к gRPC-серверу.
type Config struct {
	ServerAddress  string        // Адрес сервера (host:port)
	UseTLS         bool          // Использовать TLS
	TLSSkipVerify  bool          // Пропускать проверку сертификата
	CACertPath     string        // Путь к CA сертификату
	ConnectTimeout time.Duration // Таймаут подключения
}

// ConnManager определяет интерфейс для менеджера подключений к gRPC серверу.
//
// Метод Connect устанавливает соединение с сервером и возвращает
// активное gRPC-соединение в виде интерфейса GrpcConn,
// который оборачивает *grpc.ClientConn и предоставляет необходимые методы,
// либо ошибку, если соединение не удалось установить.
type ConnManager interface {
	Connect(ctx context.Context) (GrpcConn, error)
	Close() error
}

// GrpcConn описывает минимальный набор методов, которые мы используем от *grpc.ClientConn.
type GrpcConn interface {
	grpc.ClientConnInterface
	GetState() connectivity.State
	Close() error
	Connect()
	WaitForStateChange(ctx context.Context, s connectivity.State) bool
}

// Manager управляет состоянием и установкой gRPC-подключения.
// Он обеспечивает потокобезопасное использование соединения и автоматическое переподключение.
type Manager struct {
	config *Config
	conn   GrpcConn
	mu     sync.RWMutex
	logger *zap.Logger

	dialFunc    func(target string, opts ...grpc.DialOption) (GrpcConn, error)
	authManager auth.AuthManagerIface // интерфейс для доступа к токену
}

type connectionResult struct {
	conn GrpcConn
	err  error
}

// New создаёт новый экземпляр Manager с заданной конфигурацией и логгером.
func New(cfg *Config, logger *zap.Logger, authManager auth.AuthManagerIface) *Manager {
	return &Manager{
		config:      cfg,
		logger:      logger,
		dialFunc:    defaultDialFunc,
		authManager: authManager,
	}
}

// Connect устанавливает или повторно использует gRPC-соединение.
// При повторном вызове повторно использует текущее соединение, если оно активно.
// Иначе создаёт новое. Поддерживает таймаут через контекст.
func (m *Manager) Connect(ctx context.Context) (GrpcConn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn != nil && m.conn.GetState() == connectivity.Ready {
		m.logger.Debug("Reusing existing gRPC connection")
		return m.conn, nil
	}

	if m.conn != nil {
		m.logger.Debug("Closing stale gRPC connection")
		_ = m.conn.Close()
	}

	if m.config.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, m.config.ConnectTimeout)
		defer cancel()
	}

	resultChan := make(chan connectionResult, 1)

	go func() {
		conn, err := m.createConnection(ctx)
		resultChan <- connectionResult{conn, err}
	}()

	select {
	case <-ctx.Done():
		m.logger.Warn("gRPC connection timeout or context canceled", zap.Error(ctx.Err()))
		return nil, ctx.Err()
	case result := <-resultChan:
		if result.err != nil {
			m.logger.Error("gRPC connection failed", zap.Error(result.err))
			return nil, result.err
		}
		m.logger.Info("gRPC connection established", zap.String("server", m.config.ServerAddress))
		m.conn = result.conn
		return result.conn, nil
	}
}

// createConnection выполняет реальную установку соединения с gRPC-сервером.
// Учитывает параметры TLS, таймаут и состояние соединения.
// Используется внутри Connect и не предназначен для прямого вызова.
func (m *Manager) createConnection(ctx context.Context) (GrpcConn, error) {
	var dialOpts []grpc.DialOption

	if m.config.UseTLS {
		host, _, err := net.SplitHostPort(m.config.ServerAddress)
		if err != nil {
			host = m.config.ServerAddress
		}

		tlsConfig := &tls.Config{
			InsecureSkipVerify: m.config.TLSSkipVerify,
			ServerName:         host,
		}

		if m.config.CACertPath != "" {
			caCert, err := os.ReadFile(m.config.CACertPath)
			if err != nil {
				m.logger.Error("Failed to read CA certificate", zap.String("path", m.config.CACertPath), zap.Error(err))
				return nil, err
			}
			//m.logger.Debug("caCert", zap.String("caCert", string(caCert)))

			certPool := x509.NewCertPool()
			if !certPool.AppendCertsFromPEM(caCert) {
				m.logger.Error("Failed to add CA certificate to pool")
				return nil, errors.New("failed to add CA certificate")
			}
			tlsConfig.RootCAs = certPool

			m.logger.Debug("TLS ServerName", zap.String("ServerName", tlsConfig.ServerName))

		}

		/*
			// Диагностическая проверка TLS перед gRPC
			if err := m.testTLSConnection(tlsConfig); err != nil {
				m.logger.Error("TLS handshake failed", zap.Error(err))
				return nil, err
			}
		*/

		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
		m.logger.Debug("Using TLS for gRPC connection")
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		m.logger.Debug("Using insecure gRPC connection")
	}

	dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(
		AuthUnaryInterceptor(m.authManager, m.logger),
	))

	m.logger.Debug("Dialing gRPC server", zap.String("address", m.config.ServerAddress))
	conn, err := m.dialFunc(m.config.ServerAddress, dialOpts...)
	if err != nil {
		m.logger.Error("Failed to dial gRPC server", zap.Error(err))
		return nil, err
	}

	conn.Connect()
	state := conn.GetState()
	m.logger.Debug("Initial gRPC connection state", zap.String("state", state.String()))

	if state == connectivity.Ready {
		return conn, nil
	}

	for {
		if !conn.WaitForStateChange(ctx, state) {
			m.logger.Warn("State change wait interrupted", zap.String("state", state.String()))
			conn.Close()
			return nil, ctx.Err()
		}

		state = conn.GetState()
		m.logger.Debug("New gRPC connection state", zap.String("state", state.String()))

		switch state {
		case connectivity.Ready:
			return conn, nil
		case connectivity.TransientFailure:
			m.logger.Error("gRPC connection transient failure")
			conn.Close()
			return nil, errors.New("connection failed")
		case connectivity.Shutdown:
			m.logger.Error("gRPC connection shutdown")
			return nil, errors.New("connection shutdown")
		}
	}
}

// IsReady возвращает true, если текущее соединение установлено и находится в состоянии Ready.
func (m *Manager) IsReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conn != nil && m.conn.GetState() == connectivity.Ready
}

// Close закрывает активное gRPC-соединение, если оно было установлено.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn != nil {
		m.logger.Info("Closing gRPC connection")
		return m.conn.Close()
	}
	return nil
}

// Conn возвращает текущий *grpc.ClientConn без проверки его состояния.
// Может вернуть nil, если соединение ещё не было установлено.
func (m *Manager) Conn() GrpcConn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conn
}

/*
// testTLSConnection выполняет прямое TLS-соединение и печатает причину сбоя
func (m *Manager) testTLSConnection(tlsConfig *tls.Config) error {
	_, _, err := net.SplitHostPort(m.config.ServerAddress)
	if err != nil {
		return fmt.Errorf("invalid server address: %w", err)
	}

	conn, err := m.tlsDialFunc("tcp", m.config.ServerAddress, tlsConfig)
	if err != nil {
		// Ошибки верификации сертификата
		var certErr x509.CertificateInvalidError
		if errors.As(err, &certErr) {
			return fmt.Errorf("certificate invalid: %v", certErr)
		}

		var unknownCA x509.UnknownAuthorityError
		if errors.As(err, &unknownCA) {
			return fmt.Errorf("unknown CA: %v", unknownCA)
		}

		var hostnameErr x509.HostnameError
		if errors.As(err, &hostnameErr) {
			return fmt.Errorf("hostname mismatch: %v", hostnameErr)
		}

		return fmt.Errorf("TLS dial error: %w", err)
	}
	defer conn.Close()

	state := conn.ConnectionState()
	for i, cert := range state.PeerCertificates {
		var ipAddresses []string
		for _, ip := range cert.IPAddresses {
			ipAddresses = append(ipAddresses, ip.String())
		}

		m.logger.Debug("Peer certificate",
			zap.Int("index", i),
			zap.String("CN", cert.Subject.CommonName),
			zap.Strings("DNS SANs", cert.DNSNames),
			zap.Strings("IP SANs", ipAddresses),
			zap.Time("NotBefore", cert.NotBefore),
			zap.Time("NotAfter", cert.NotAfter))
	}

	return nil
}
*/

func defaultDialFunc(target string, opts ...grpc.DialOption) (GrpcConn, error) {
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
