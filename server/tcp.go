package server

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/trialcyber/hfrpc/common"
	"github.com/trialcyber/hfrpc/interceptor"
	"github.com/trialcyber/hfrpc/packer"
)

const (
	BufferSize = 1024
)

type Tcp struct {
	Ip         string
	Port       string
	Server     common.Server
	BufferSize int
}

func NewTcpServer(ip string, port string) *Tcp {
	return &Tcp{
		ip,
		port,
		common.Server{},
		BufferSize,
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
	p.Server.Packer = &packer
}
func (p *Tcp) GetIp() string {
	return p.Ip
}
func (p *Tcp) GetPort() string {
	return p.Port
}

func (p *Tcp) handleFunc(conn net.Conn) {
	buf := make([]byte, p.BufferSize)
	go time.AfterFunc(61*time.Second, func() {
		_ = conn.Close()
	})
	for {
		n, _ := conn.Read(buf)
		if n > 0 {
			data := buf[:n]
			res := p.Server.Handler(data)
			if pack := *(p.Server.Packer); pack != nil {
				res = pack.Pack(res)
			}
			_, _ = conn.Write(res)
			buf = make([]byte, p.BufferSize)
		}
	}

}
