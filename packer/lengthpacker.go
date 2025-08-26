package packer

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
)

type LengthPacker struct {
	Offset uint32
}

func (l LengthPacker) Pack(data []byte) []byte {
	msgLen := uint32(len(string(data)))
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, msgLen)
	if err != nil {
		return nil
	}
	buf.Write(data)
	return buf.Bytes()
}

func (l LengthPacker) Unpack(conn net.Conn) ([]byte, error) {
	lenBytes := make([]byte, l.Offset)
	_, err := io.ReadFull(conn, lenBytes)
	if err != nil {
		return nil, err
	}
	// 解析消息长度
	msgLength := binary.BigEndian.Uint32(lenBytes)
	// 读取消息体
	msgBytes := make([]byte, msgLength)
	_, err = io.ReadFull(conn, msgBytes)
	if err != nil {
		return nil, err
	}
	return msgBytes, nil
}
