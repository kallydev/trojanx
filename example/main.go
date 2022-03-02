package main

import (
	"context"
	"crypto/tls"
	"github.com/kallydev/trojanx"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"net/http"
	"time"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})
}

func main() {
	go func() {
		server := &http.Server{
			Addr:         "127.0.0.1:80",
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		}
		server.SetKeepAlivesEnabled(false)
		http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			defer request.Body.Close()
			logrus.Debugln(request.RemoteAddr, request.RequestURI)
			host, _, _ := net.SplitHostPort(request.Host)
			switch host {
			default:
				writer.Header().Set("Connection", "close")
				writer.Header().Set("Referrer-Policy", "no-referrer")
				http.Redirect(writer, request, "https://example.com/", http.StatusFound)
			}
		})
		if err := server.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()
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
		return true
	}
	srv.AuthenticationHandler = func(ctx context.Context, hash string) bool {
		switch hash {
		// TODO verify password
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
