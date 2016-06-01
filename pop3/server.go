package pop3

import (
	log "github.com/sirupsen/logrus"
	"github.com/trapped/gomaild2/pop3/locker"
	. "github.com/trapped/gomaild2/pop3/structs"
	. "github.com/trapped/gomaild2/structs"
	"net"
	"time"
)

type Server struct {
	Addr    string
	Port    string
	Timeout time.Duration
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
		if c.GetBool("authenticated") {
			locker.Unlock(c.GetString("authenticated_as"))
		}
		c.Conn.Close()
		log.Info("Disconnected")
	}()

	c.Send(Reply{Result: OK, Message: "gomaild2 POP3 ready"})
	c.State = Authorization
	log.Info("Connected")

	// anonymous struct chan for catching a returned command and error
	// basically an adapter for Client.Receive(), used for timeouts
	recv := make(chan struct {
		Command
		error
	}, 1)

cmdloop:
	for {
		if c.State == Disconnected {
			return
		}

		// get the command from the user
		go func() {
			cmd, err := c.Receive()
			recv <- struct {
				Command
				error
			}{cmd, err}
		}()

		// if no commands sent in x min, kill session
		select {
		case r := <-recv:
			if r.error != nil {
				break cmdloop
			}
			c.Send(Process(c, r.Command))

		case <-time.After(s.Timeout):
			c.Send(Reply{Result: ERR, Message: "session timeout"})
			return
		}
	}
}
