package trojanx

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/kallydev/trojanx/internal/pipe"
	"github.com/kallydev/trojanx/metadata"
	"github.com/kallydev/trojanx/protocol"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
)

type Server struct {
	ctx         context.Context
	config      *Config
	tcpListener net.Listener
	tlsListener net.Listener

	// TODO some callback functions
	ConnectHandler        ConnectHandler
	AuthenticationHandler AuthenticationHandler
	RequestHandler        RequestHandler
	ErrorHandler          ErrorHandler
	// TODO add a record callback handler
}

func (s *Server) run() error {
	var err error
	s.tcpListener, err = net.Listen("tcp", net.JoinHostPort(s.config.Host, strconv.Itoa(s.config.Port)))
	if err != nil {
		return err
	}
	var tlsCertificates []tls.Certificate
	if s.config.TLSConfig != nil {
		for _, certificateFile := range s.config.TLSConfig.CertificateFiles {
			certificate, err := tls.LoadX509KeyPair(certificateFile.PublicKeyFile, certificateFile.PrivateKeyFile)
			if err != nil {
				return err
			}
			tlsCertificates = append(tlsCertificates, certificate)
		}
		s.tlsListener = tls.NewListener(s.tcpListener, &tls.Config{
			Certificates: tlsCertificates,
		})
	}
	for {
		var conn net.Conn
		if s.tlsListener == nil {
			conn, err = s.tcpListener.Accept()
		} else {
			conn, err = s.tlsListener.Accept()
		}
		if err != nil {
			s.ErrorHandler(s.ctx, err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	// TODO Not used for now
	ctx := metadata.NewContext(context.Background(), metadata.Metadata{
		LocalAddr:  conn.LocalAddr(),
		RemoteAddr: conn.RemoteAddr(),
	})
	if !s.ConnectHandler(ctx) {
		return
	}
	token, err := protocol.GetToken(conn)
	if err != nil && token == "" {
		s.ErrorHandler(ctx, err)
		return
	}
	if !s.AuthenticationHandler(ctx, token) {
		logrus.Debugln("authentication not passed", conn.RemoteAddr())
		if s.config.ReverseProxyConfig == nil {
			return
		}
		remoteURL := net.JoinHostPort(s.config.ReverseProxyConfig.Host, strconv.Itoa(s.config.ReverseProxyConfig.Port))
		dst, err := net.Dial("tcp", remoteURL)
		if err != nil {
			s.ErrorHandler(ctx, err)
			return
		}
		logrus.Debugln("reverse proxy policy", conn.RemoteAddr(), dst.LocalAddr())
		defer dst.Close()
		if _, err := dst.Write([]byte(token)); err != nil {
			s.ErrorHandler(ctx, err)
			return
		}
		go pipe.Copy(dst, conn)
		pipe.Copy(conn, dst)
		return
	}
	req, err := protocol.ParseRequest(conn)
	if err != nil {
		s.ErrorHandler(ctx, err)
		return
	}
	if req.Command == protocol.CommandUDP {
		s.ErrorHandler(ctx, errors.New("unsupported udp protocol"))
		return
	}
	dst, err := net.Dial("tcp", net.JoinHostPort(req.DescriptionAddress, strconv.Itoa(req.DescriptionPort)))
	if err != nil {
		s.ErrorHandler(ctx, err)
		return
	}
	defer dst.Close()
	go pipe.Copy(dst, conn)
	pipe.Copy(conn, dst)
}

func (s *Server) Run() error {
	errCh := make(chan error)
	go func() {
		errCh <- s.run()
	}()
	select {
	case err := <-errCh:
		return err
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

func New(ctx context.Context, config *Config) *Server {
	return &Server{
		ctx:                   ctx,
		config:                config,
		ConnectHandler:        DefaultConnectHandler,
		AuthenticationHandler: DefaultAuthenticationHandler,
		RequestHandler:        DefaultRequestHandler,
		ErrorHandler:          DefaultErrorHandler,
	}
}
