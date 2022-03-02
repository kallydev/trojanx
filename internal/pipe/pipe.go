package pipe

import (
	"errors"
	"github.com/kallydev/trojanx/internal/pool"
	"io"
	"net"
)

func Copy(dst net.Conn, src net.Conn) (written int64, err error) {
	defer dst.Close()
	buffer := pool.Get()
	defer pool.Put(buffer)
	for {
		n, err := src.Read(buffer)
		if n > 0 {
			n, err := dst.Write(buffer[:n])
			if err != nil {
				if errors.Is(err, io.EOF) {
					return written, nil
				}
				return written, err
			}
			written += int64(n)
		}
		if err != nil {
			return written, err
		}
	}
}
