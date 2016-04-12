package smtp

import (
	"bufio"
	"net"
	. "trapped/gomaild2/smtp/structs"
)

type Server struct {
	Addr string
	Port string
}

//go:generate gengen -d ./commands/ -t process.go.tmpl -o process.go

func (s *Server) Start() {
	l, err := net.Listen("tcp", s.Addr+":"+s.Port)
	if err != nil {
		panic(err.Error())
	}
	for {
		c, _ := l.Accept()
		client := &Client{
			Conn: c,
			Rdr:  bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c)),
			Data: make(map[string]interface{}),
		}
		go accept(client)
	}
}

func accept(c *Client) {
	c.Send(Reply{Result: Ready, Message: "ready"})
	c.State = Connected
	for {
		if c.State == Disconnected {
			break
		}
		c.Send(Process(c, c.Receive()))
	}
	c.Conn.Close()
}
