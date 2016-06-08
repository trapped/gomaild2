package smtp

import (
	. "github.com/lsegal/gucumber"
	log "github.com/sirupsen/logrus"
	. "github.com/trapped/gomaild2/smtp"
	. "github.com/trapped/gomaild2/smtp/structs"
	"io/ioutil"
	"strings"
	"time"
)

func createClient(addr string) {
	World["client"] = &Client{}
	err := World["client"].(*Client).Connect(addr)
	if err != nil {
		T.Errorf("client failed to connect: %s", err)
	}
}

func clientExpectReceive(code string) Command {
	reply, err := World["client"].(*Client).Receive()
	if err != nil {
		T.Errorf("client failed to receive reply: %s", err)
	}
	if reply.Verb != code {
		T.Errorf("reply code mismatch: %v (expected %v)", reply.Verb, code)
	}
	return reply
}

func init() {
	Before("@smtp", func() {
		log.SetOutput(ioutil.Discard)
		World = make(map[string]interface{})
	})
	After("@smtp", func() {
		if s, ok := World["server"]; ok {
			s.(*Server).Stop <- true
		}
		if c, ok := World["client"]; ok {
			c.(*Client).Conn.Close()
		}
		World = nil
	})

	Given(`^a server is listening on "(.+)"$`, func(addr string) {
		s := strings.Split(addr, ":")
		World["server"] = &Server{
			Addr:    s[0],
			Port:    s[1],
			Timeout: 1 * time.Minute,
		}
		err := World["server"].(*Server).Start()
		if err != nil {
			T.Errorf("failed to start server: %v", err)
		}
	})

	Given(`^a client is connected to "(.+)"$`, func(addr string) {
		createClient(addr)
		clientExpectReceive("220")
	})

	When(`^a client connects to "(.+)"$`, func(addr string) {
		createClient(addr)
	})

	Then(`^the client should receive a ([0-9]{3}) reply$`, func(code string) {
		clientExpectReceive(code)
	})

	When(`^the client sends a "(.+)" command with args "(.+)"$`, func(verb, args string) {
		World["client"].(*Client).SendCmd(Command{
			Verb: verb,
			Args: args,
		})
	})
}
