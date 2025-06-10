package packer

import (
	"encoding/binary"
)

type LengthPacker struct {
	Offset uint32
}

func (l LengthPacker) Pack(data []byte) []byte {
	buf := make([]byte, int(l.Offset)+len(data))
	binary.BigEndian.PutUint32(buf[:l.Offset], uint32(len(data)))
	copy(buf[l.Offset:], data)
	return buf
}

func (l LengthPacker) Unpack(data []byte) []byte {
	dataLen := binary.BigEndian.Uint32(data[:l.Offset])
	return data[l.Offset : dataLen+l.Offset]
}
