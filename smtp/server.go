package smtp

import (
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	. "github.com/trapped/gomaild2/smtp/structs"
	. "github.com/trapped/gomaild2/structs"
	"net"
	"time"
)

type Server struct {
	Addr        string
	Port        string
	RequireAuth bool
	Outbound    bool
	Timeout     time.Duration
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

	// set a timeout for the inbound server
	if !s.Outbound {
		s.Timeout = time.Duration(config.GetInt("server.smtp.mta.timeout")) * time.Second
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
		go s.accept(client)
	}
}

func (s *Server) accept(c *Client) {
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

	c.Send(Reply{Result: Ready, Message: config.GetString("server.name") + " gomaild2 ESMTP ready"})
	c.State = Connected
	log.Info("Connected")

	recv := make(chan struct {
		Command
		error
	}, 1)

	outbound := s.Outbound
cmdloop:
	for {
		if c.State == Disconnected {
			break
		}
		// only timeout on !outbound smtp clients
		if !outbound {
			go func() {
				cmd, err := c.Receive()
				recv <- struct {
					Command
					error
				}{cmd, err}
			}()

			select {
			case r := <-recv:
				if r.error != nil {
					break cmdloop
				}
				c.Send(Process(c, r.Command))

			case <-time.After(s.Timeout):
				c.Send(Reply{Result: Closing, Message: "session timeout"})
				return
			}

		} else {
			cmd, err := c.Receive()
			if err != nil {
				break
			}
			c.Send(Process(c, cmd))
		}
	}
}
