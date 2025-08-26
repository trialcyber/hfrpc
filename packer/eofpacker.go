package packer

import (
	"io"
	"net"
)

type Packer interface {
	Pack([]byte) []byte
	Unpack(conn net.Conn) ([]byte, error)
}

type EofPacker struct {
	Eof string
}

func (e EofPacker) Pack(data []byte) []byte {
	return append(data, []byte(e.Eof)...)
}

func (e EofPacker) Unpack(conn net.Conn) ([]byte, error) {
	msgBytes := make([]byte, 1024*1024)
	n, err := io.ReadFull(conn, msgBytes)
	if err != nil {
		return nil, err
	}
	data := msgBytes[:n]
	return msgBytes[:len(data)-len([]byte(e.Eof))], nil
}
