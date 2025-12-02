package hfrpc

import (
	"errors"

	"github.com/trialcyber/hfrpc/interceptor"
	"github.com/trialcyber/hfrpc/packer"
	"github.com/trialcyber/hfrpc/server"
)

type ServerInterface interface {
	Start()
	Register(s interface{})
	SetBuffer(bs int)
	GetIp() string
	GetPort() string
	SetPacker(packer packer.Packer)
	SetInterceptor(handlers []interceptor.HandlerFunc)
}

func NewServer(protocol string, ip string, port string) (ServerInterface, error) {
	var err error
	switch protocol {
	case "http":
		return server.NewHttpServer(ip, port), err
	case "tcp":
		return server.NewTcpServer(ip, port), err
	}
	return nil, errors.New("The protocol can not be supported")
}
