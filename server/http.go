package server

import (
	"fmt"
	"github.com/trialcyber/hfrpc/common"
	"github.com/trialcyber/hfrpc/interceptor"
	"github.com/trialcyber/hfrpc/packer"
	"io"
	"log"
	"net/http"
)

type Http struct {
	Ip         string
	Port       string
	Server     common.Server
	BufferSize int
	Packer     *packer.Packer
}

func NewHttpServer(ip string, port string) *Http {
	return &Http{
		ip,
		port,
		common.Server{},
		BufferSize,
		nil,
	}
}

func (p *Http) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handleFunc)
	var url = fmt.Sprintf("%s:%s", p.Ip, p.Port)
	log.Printf("Listening http://%s:%s", p.Ip, p.Port)
	_ = http.ListenAndServe(url, mux)
}

func (p *Http) Register(s interface{}) {
	p.Server.Register(s)
}

func (p *Http) SetBuffer(bs int) {
	p.BufferSize = bs
}

func (p *Http) SetInterceptor(handlers []interceptor.HandlerFunc) {
	p.Server.Interceptor = handlers
}

func (p *Http) SetPacker(packer packer.Packer) {
	p.Packer = &packer
}

func (p *Http) GetIp() string {
	return p.Ip
}

func (p *Http) GetPort() string {
	return p.Port
}

func (p *Http) handleFunc(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		data []byte
	)
	w.Header().Set("Content-Type", "application/json")
	if data, err = io.ReadAll(r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	res := p.Server.Handler(data)
	_, _ = w.Write(res)
}
