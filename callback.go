package trojanx

import (
	"context"
	"github.com/kallydev/trojanx/protocol"
	"github.com/sirupsen/logrus"
	"net"
)

type (
	ConnectHandler        = func(ctx context.Context) bool
	AuthenticationHandler = func(ctx context.Context, hash string) bool
	RequestHandler        = func(ctx context.Context, request protocol.Request) bool
	ForwardHandler        = func(ctx context.Context, upload, download int64) bool
	ErrorHandler          = func(ctx context.Context, err error)
)

func DefaultConnectHandler(ctx context.Context) bool {
	return true
}

func DefaultAuthenticationHandler(ctx context.Context, hash string) bool {
	return false
}

func DefaultRequestHandler(ctx context.Context, request protocol.Request) bool {
	var remoteIP net.IP
	if request.AddressType == protocol.AddressTypeDomain {
		tcpAddr, err := net.ResolveTCPAddr("tcp", request.DescriptionAddress)
		if err != nil {
			logrus.Errorln(err)
			return false
		}
		remoteIP = tcpAddr.IP
	} else {
		remoteIP = net.ParseIP(request.DescriptionAddress)
	}
	if remoteIP.IsLoopback() || remoteIP.IsLinkLocalUnicast() || remoteIP.IsLinkLocalMulticast() || remoteIP.IsPrivate() {
		return false
	}
	return true
}

func DefaultErrorHandler(ctx context.Context, err error) {
	logrus.Errorln(err)
}
