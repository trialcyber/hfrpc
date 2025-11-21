package server

import (
	"fmt"
	"io"
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
	pack := *(p.Server.Packer)
	defer conn.Close()
	for {
		data, err := pack.Unpack(conn)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("%v read from client failed, err: %v \n",
					time.Now().Format("2006-01-02 15:04:05"),
					err.Error(),
				)
			}
			break
		}
		fmt.Println(string(data))
		res := p.Server.Handler(data)
		res = pack.Pack(res)
		if res != nil {
			_, _ = conn.Write(res)
		}
	}

}
