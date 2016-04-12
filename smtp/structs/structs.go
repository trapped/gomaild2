package smtp

import (
	"bufio"
	"net"
	"strconv"
	"strings"
)

type StatusCode int

const (
	Ready                 StatusCode = 220
	Closing                          = 221
	Ok                               = 250
	Unavailable                      = 421
	CommandSyntaxError               = 500
	ArgumentSyntaxError              = 501
	AlreadyIdentified                = 503
	CommandNotImplemented            = 504
)

type SessionState int

const (
	Any SessionState = iota
	Connected
	Identified       //after EHLO before MAIL
	ReceivingHeaders //after MAIL before RCPT
	ReceivingBody    //after DATA
	Disconnected
)

type Command struct {
	Verb string
	Args string
}

type Reply struct {
	Result  StatusCode
	Message string
}

type Client struct {
	Conn  net.Conn
	State SessionState
	Rdr   *bufio.ReadWriter
	Data  map[string]interface{}
}

func (c *Client) Send(r Reply) {
	_, err := c.Rdr.WriteString(strconv.Itoa(int(r.Result)) + " " + r.Message + "\r\n")
	c.Rdr.Flush()
	if err != nil {
		c.Conn.Close()
		c.State = Disconnected
	}
}

func (c *Client) Receive() Command {
	str, err := c.Rdr.ReadString('\n')
	str = strings.TrimSpace(str)
	if err != nil {
		c.Conn.Close()
		c.State = Disconnected
	}
	split := strings.Split(str, " ")
	return Command{
		Verb: split[0],
		Args: strings.Join(split[1:], " "),
	}
}
