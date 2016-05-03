package smtp

import (
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	. "github.com/trapped/gomaild2/smtp/structs"
	. "github.com/trapped/gomaild2/structs"
	"net"
)

type Server struct {
	Addr        string
	Port        string
	RequireAuth bool
	Outbound    bool
}

//go:generate gengen -d ./commands/ -t process.go.tmpl -o process.go

func (s *Server) Start() {
	log.WithFields(log.Fields{
		"addr":         s.Addr,
		"port":         s.Port,
		"outbound":     s.Outbound,
		"require_auth": s.RequireAuth,
	}).Info("Starting SMTP server")
	l, err := net.Listen("tcp", s.Addr+":"+s.Port)
	if err != nil {
		panic(err.Error())
	}
	for {
		c, _ := l.Accept()
		client := &Client{
			Conn: c,
			Data: make(map[string]interface{}),
			ID:   SessionID(12),
		}
		client.MakeReader()
		client.Set("outbound", s.Outbound)
		client.Set("require_auth", s.RequireAuth)
		client.SaveData()
		go accept(client)
	}
}

func accept(c *Client) {
	defer func() {
		c.Conn.Close()
		log.WithFields(log.Fields{
			"id":   c.ID,
			"addr": c.Conn.RemoteAddr().String(),
		}).Info("Disconnected")
	}()

	c.Send(Reply{Result: Ready, Message: config.GetString("server.name") + " gomaild2 ESMTP ready"})
	c.State = Connected
	log.WithFields(log.Fields{
		"id":   c.ID,
		"addr": c.Conn.RemoteAddr().String(),
	}).Info("Connected")

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
