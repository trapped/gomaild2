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
	Stop        chan bool
}

//go:generate gengen -d ./commands/ -t process.go.tmpl -o process.go

func (s *Server) MakeClient(c net.Conn) *Client {
	client := &Client{
		Conn: c,
		Data: make(map[string]interface{}),
		ID:   SessionID(12),
	}
	client.MakeReader()
	client.Set("outbound", s.Outbound)
	client.Set("require_auth", s.RequireAuth)
	client.SaveData()
	return client
}

func (s *Server) Start() error {
	log.WithFields(log.Fields{
		"addr":         s.Addr,
		"port":         s.Port,
		"outbound":     s.Outbound,
		"require_auth": s.RequireAuth,
	}).Info("Starting SMTP server")
	addr, err := net.ResolveTCPAddr("tcp", s.Addr+":"+s.Port)
	if err != nil {
		log.Errorf("Failed to resolve listen address '%s': %v", s.Addr+":"+s.Port, err)
		return err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Errorf("Failed to listen on address '%s': %v", s.Addr+":"+s.Port, err)
		return err
	}

	s.Stop = make(chan bool)
	go func() {
		for {
			select {
			case <-s.Stop:
				log.Info("Stopping SMTP server...")
				l.Close()
				return
			default:
			}
			//prepare async accept()
			l.SetDeadline(time.Now().Add(100 * time.Millisecond))
			c, err := l.Accept()
			if err == nil {
				client := s.MakeClient(c)
				go s.accept(client)
			}
		}
	}()
	return nil
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

cmdloop:
	for {
		if c.State == Disconnected {
			break
		}

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
	}
}
