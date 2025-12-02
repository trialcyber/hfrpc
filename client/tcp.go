package client

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/trialcyber/hfrpc/common"
	"github.com/trialcyber/hfrpc/packer"
)

var Connection map[string]ConnInfo

type ConnInfo struct {
	Addr        string
	Conn        net.Conn
	LastUseTime int64
}

type Tcp struct {
	Ip      string
	Port    string
	Options OptionsSetting
	Packer  *packer.Packer
}

type OptionsSetting struct {
	Heartbeat   int
	MaxIdleTime int
}

func (p *Tcp) Call(method string, params interface{}, result interface{}) (err error) {
	req := common.JsonRs(strconv.FormatInt(time.Now().Unix(), 10), method, params)
	err = p.handleFunc(req, result)
	return err
}

func (p *Tcp) handleFunc(b []byte, result interface{}) error {
	var addr = fmt.Sprintf("%s:%s", p.Ip, p.Port)
	conn, err := p.GetConnection(addr)
	if err != nil {
		return err
	}

	if p.Packer != nil { // 组装数据包
		b = (*(p.Packer)).Pack(b)
	}
	_, err = conn.Write(b)
	if err != nil {
		p.CloseConnection(addr, conn)
		return err
	}

	//解析数据包
	var res []byte
	if p.Packer != nil {
		res, err = (*(p.Packer)).Unpack(conn)
	} else {
		var buf = make([]byte, 512)
		n, _ := conn.Read(buf)
		res = buf[:n]
	}
	if err != nil {
		p.CloseConnection(addr, conn)
		return err
	}
	err = common.GetResult(res, result)
	return err
}

func (p *Tcp) GetConnection(addr string) (con net.Conn, err error) {
	timeUnix := time.Now().Unix()
	if res, ok := Connection[addr]; ok { // 命中内存的连接
		res.LastUseTime = timeUnix
		Connection[addr] = res
		return res.Conn, nil
	}
	// 首次创建连接
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	if Connection == nil {
		Connection = make(map[string]ConnInfo)
	}
	// 存入变量
	Connection[addr] = ConnInfo{Addr: addr, Conn: conn, LastUseTime: timeUnix}
	// 检测心跳 和 最大空闲时间
	go func() {
		for {
			heartbeat := 7
			if p.Options.Heartbeat > 0 {
				heartbeat = p.Options.Heartbeat
			}
			time.Sleep(time.Duration(heartbeat) * time.Second)
			_, err := conn.Write([]byte("ping")) //检测心跳
			//最大空闲时间
			ti := time.Now().Unix() - Connection[addr].LastUseTime
			maxIdleTime := 60
			if p.Options.MaxIdleTime > 0 {
				maxIdleTime = p.Options.MaxIdleTime
			}
			if err != nil || ti > int64(maxIdleTime) {
				p.CloseConnection(addr, conn)
				break
			}
		}
	}()
	return conn, err
}

func (*Tcp) CloseConnection(addr string, conn net.Conn) {
	conn.Close()
	delete(Connection, addr)
}
