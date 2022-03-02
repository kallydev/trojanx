package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/kallydev/trojanx/internal/pool"
	"io"
	"net"
)

const (
	LenToken       = 56
	LenCRLF        = 2
	LenCommand     = 1
	LenAddressType = 1
	LenIPv4        = 4
	LenDomain      = 1
	LenIPv6        = 16
	LenPort        = 2

	CommandConnect = 0x01
	CommandUDP     = 0x03

	AddressTypeIPv4   = 0x01
	AddressTypeDomain = 0x03
	AddressTypeIPv6   = 0x04
)

var (
	CRLF = []byte{'\x0D', '\x0A'}
)

func GetToken(conn net.Conn) (string, error) {
	buffer := pool.Get()
	defer pool.Put(buffer)
	if n, err := conn.Read(buffer[:LenToken]); err != nil || n != LenToken {
		if err != nil {
			return string(buffer[:LenToken]), err
		}
		return string(buffer[:LenToken]), errors.New("token length error")
	}
	return string(buffer[:LenToken]), nil
}

type Request struct {
	Command            byte
	AddressType        byte
	DescriptionAddress string
	DescriptionPort    int
}

func ParseRequest(conn net.Conn) (*Request, error) {
	request := &Request{}
	buffer := pool.Get()
	defer pool.Put(buffer)
	if _, err := io.ReadFull(conn, buffer[:LenCRLF]); err != nil {
		return nil, err
	}
	if !bytes.Equal(buffer[:LenCRLF], CRLF) {
		return nil, errors.New("not is crlf data")
	}
	if _, err := io.ReadFull(conn, buffer[:LenCommand]); err != nil {
		return nil, err
	}
	request.Command = buffer[LenCommand-1]
	if _, err := io.ReadFull(conn, buffer[:LenAddressType]); err != nil {
		return nil, err
	}
	request.AddressType = buffer[LenAddressType-1]
	switch request.AddressType {
	case AddressTypeIPv4:
		if _, err := io.ReadFull(conn, buffer[:LenIPv4]); err != nil {
			return nil, err
		}
		request.DescriptionAddress = net.IP(buffer[:LenIPv4]).String()
	case AddressTypeDomain:
		if _, err := io.ReadFull(conn, buffer[:LenDomain]); err != nil {
			return nil, err
		}
		l := buffer[LenDomain-1]
		if _, err := io.ReadFull(conn, buffer[:l]); err != nil {
			return nil, err
		}
		request.DescriptionAddress = string(buffer[:l])
	case AddressTypeIPv6:
		if _, err := io.ReadFull(conn, buffer[:LenIPv6]); err != nil {
			return nil, err
		}
		request.DescriptionAddress = net.IP(buffer[:LenIPv6]).String()
	default:
		return nil, fmt.Errorf("unsupported address type %d", request.AddressType)
	}
	if _, err := io.ReadFull(conn, buffer[:LenPort]); err != nil {
		return nil, err
	}
	request.DescriptionPort = int(binary.BigEndian.Uint16(buffer[:LenPort]))
	if _, err := io.ReadFull(conn, buffer[:LenCRLF]); err != nil {
		return nil, err
	}
	if !bytes.Equal(buffer[:LenCRLF], CRLF) {
		return nil, errors.New("not is crlf data")
	}
	return request, nil
}
