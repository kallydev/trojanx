# TrojanX

Trojan server **framework**.

## Example

```go
package main

import (
	"context"
	"crypto/tls"
	"github.com/kallydev/trojanx"
	"github.com/sirupsen/logrus"
	"net"
)

func main() {
	srv := trojanx.New(context.Background(), &trojanx.Config{
		Host: net.IPv4zero.String(),
		Port: 443,
		TLSConfig: &trojanx.TLSConfig{
			MinVersion: tls.VersionTLS13,
			MaxVersion: tls.VersionTLS13,
			CertificateFiles: []trojanx.CertificateFileConfig{
				{
					PublicKeyFile:  "/etc/letsencrypt/live/example.com/fullchain.pem",
					PrivateKeyFile: "/etc/letsencrypt/live/example.com/privkey.pem",
				},
			},
		},
		ReverseProxyConfig: &trojanx.ReverseProxyConfig{
			Scheme: "http",
			Host:   "127.0.0.1",
			Port:   80,
		},
	})
	srv.ConnectHandler = func(ctx context.Context) bool {
		// TODO Allow or deny connection requests
		return true
	}
	srv.AuthenticationHandler = func(ctx context.Context, hash string) bool {
		switch hash {
		// TODO Verify password
		default:
			return false
		}
	}
	srv.ErrorHandler = func(ctx context.Context, err error) {
		logrus.Errorln(err)
	}
	if err := srv.Run(); err != nil {
		logrus.Fatalln(err)
	}
}
```

## LICENSE

[MIT License](LICENSE)
