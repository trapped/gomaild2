package pop3

import (
	log "github.com/sirupsen/logrus"
	. "github.com/trapped/gomaild2/pop3/structs"
	. "github.com/trapped/gomaild2/structs"
	"net"
)

type Server struct {
	Addr string
	Port string
}

//go:generate gengen -d ./commands/ -t process.go.tmpl -o process.go

func (s *Server) Start() {
	log.WithFields(log.Fields{
		"addr": s.Addr,
		"port": s.Port,
	}).Info("Starting POP3 server")
	l, err := net.Listen("tcp", s.Addr+":"+s.Port)
	if err != nil {
		panic(err.Error())
	}
	for {
		c, _ := l.Accept() // should handle this error
		client := &Client{
			Conn: c,
			Data: make(map[string]interface{}),
			ID:   SessionID(12),
		}
		client.MakeReader()
		go accept(client)
	}
}

func accept(c *Client) {
	log := log.WithFields(log.Fields{
		"id":   c.ID,
		"addr": c.Conn.RemoteAddr().String(),
	})
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
		c.Conn.Close()
		log.Info("Disconnected")
	}()

	c.Send(Reply{Result: OK, Message: "gomaild2 POP3 ready"})
	c.State = Authorization
	log.Info("Connected")

	for {
		if c.State == Disconnected {
			break
		}
		cmd, err := c.Receive()
		if err != nil {
			break
		}
		c.Send(Process(c, cmd))
	}
}
