package packer

import (
	"encoding/binary"
	"errors"
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

func (l LengthPacker) Unpack(data []byte) ([]byte, error) {
	dataLen := binary.BigEndian.Uint32(data[:l.Offset])
	if int(dataLen) > (len(data) - int(l.Offset)) {
		err := errors.New("format error")
		return data, err
	}
	return data[l.Offset : dataLen+l.Offset], nil
}
