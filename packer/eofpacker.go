package packer

type Packer interface {
	Pack([]byte) []byte
	Unpack(data []byte) ([]byte, error)
}

type EofPacker struct {
	Eof string
}

func (e EofPacker) Pack(data []byte) []byte {
	return append(data, []byte(e.Eof)...)
}

func (e EofPacker) Unpack(data []byte) ([]byte, error) {
	return data[:len(data)-len([]byte(e.Eof))], nil
}
