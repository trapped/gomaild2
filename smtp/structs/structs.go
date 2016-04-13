package smtp

import (
	"bufio"
	"net"
	"regexp"
	"strconv"
	"strings"
)

var (
	RxDomain       *regexp.Regexp = regexp.MustCompile(`^([a-zA-Z0-9]{1}[a-zA-Z0-9_-]{1,62}){1}(\.[a-zA-Z0-9]{1}[a-zA-Z0-9_-]{1,62})+$`)
	RxEmailAddress *regexp.Regexp = regexp.MustCompile("^(((([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|((\\x22)((((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(([\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(\\([\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(\\x22)))@((([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$")
)

type StatusCode int

const (
	Ready                 StatusCode = 220
	Closing                          = 221
	Ok                               = 250
	Unavailable                      = 421
	CommandSyntaxError               = 500
	ArgumentSyntaxError              = 501
	BadSequence                      = 503
	CommandNotImplemented            = 504
)

type SessionState int

const (
	Any SessionState = iota
	Connected
	Identified       //after EHLO before MAIL
	ReceivingHeaders //after MAIL before/during RCPT
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
