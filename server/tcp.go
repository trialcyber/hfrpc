package server

import (
	"fmt"
	"hfrpc/common"
	"hfrpc/interceptor"
	"hfrpc/packer"
	"log"
	"net"
)

const (
	BufferSize = 1024
)

type Tcp struct {
	Ip         string
	Port       string
	Server     common.Server
	BufferSize int
	Packer     *packer.Packer
}

func NewTcpServer(ip string, port string) *Tcp {
	return &Tcp{
		ip,
		port,
		common.Server{},
		BufferSize,
		nil,
	}
}

func (p *Tcp) Start() {
	var addr = fmt.Sprintf("%s:%s", p.Ip, p.Port)
	listener, _ := net.Listen("tcp", addr)
	log.Printf("Listening tcp://%s:%s", p.Ip, p.Port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go p.handleFunc(conn)
	}
}

func (p *Tcp) Register(s interface{}) {
	p.Server.Register(s)
}

func (p *Tcp) SetBuffer(bs int) {
	p.BufferSize = bs
}

func (p *Tcp) SetInterceptor(handlers []interceptor.HandlerFunc) {
	p.Server.Interceptor = handlers
}

func (p *Tcp) SetPacker(packer packer.Packer) {
	p.Packer = &packer
}

func (p *Tcp) handleFunc(conn net.Conn) {

	defer conn.Close()
	for {
		var buf = make([]byte, p.BufferSize)
		n, _ := conn.Read(buf)
		if n > 0 {
			data := buf[:n]
			if p.Packer != nil {
				data = (*p.Packer).Unpack(data)
			}
			res := p.Server.Handler(data)
			if p.Packer != nil {
				res = (*p.Packer).Pack(res)
			}
			_, _ = conn.Write(res)
		}
	}

}
