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
	RxIP           *regexp.Regexp = regexp.MustCompile("^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$")
	RxEmailAddress *regexp.Regexp = regexp.MustCompile("^(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\\])$")
)

type StatusCode int

const (
	Ignore                   StatusCode = -1
	Ready                               = 220
	Closing                             = 221
	AuthenticationSuccessful            = 235
	Ok                                  = 250
	StartAuth                           = 334
	StartSending                        = 354
	Unavailable                         = 421
	LocalError                          = 451
	CommandSyntaxError                  = 500
	ArgumentSyntaxError                 = 501
	BadSequence                         = 503
	CommandNotImplemented               = 504
	AuthenticationRequired              = 530
	AuthenticationInvalid               = 535
	TransactionFailed                   = 554
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

//packages should append() during init()
var Extensions []string = []string{
	"PIPELINING",
	"8BITMIME",
}

type Command struct {
	Verb string
	Args string
}

type Reply struct {
	Result  StatusCode
	Message string
}

type Client struct {
	Conn         net.Conn
	ID           string
	State        SessionState
	Rdr          *bufio.ReadWriter
	Data         map[string]interface{}
	default_data map[string]interface{}
}

func (c *Client) Connect(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	c.Conn = conn
	c.MakeReader()
	return nil
}

func (c *Client) MakeReader() {
	c.Rdr = bufio.NewReadWriter(bufio.NewReader(c.Conn), bufio.NewWriter(c.Conn))
}

func (c *Client) SaveData() {
	c.default_data = c.Data
}

func (c *Client) ResetData() {
	c.Data = c.default_data
}

func (c *Client) Set(key string, value interface{}) {
	c.Data[key] = value
}

func (c *Client) Get(key string) interface{} {
	return c.Data[key]
}

func (c *Client) GetBool(key string) bool {
	if v, ok := c.Data[key]; ok {
		return v.(bool)
	}
	return false
}

func (c *Client) GetString(key string) string {
	if v, ok := c.Data[key]; ok {
		return v.(string)
	}
	return ""
}

func (c *Client) GetStringSlice(key string) []string {
	if v, ok := c.Data[key]; ok {
		return v.([]string)
	}
	return make([]string, 0)
}

func IsSuccess(r Reply) bool {
	if int(r.Result) < 400 {
		return true
	} else {
		return false
	}
}

func FormatReply(r Reply) string {
	lines := strings.Split(r.Message, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		separator := "-"
		if i >= len(lines)-1 {
			separator = " "
		}
		lines[i] = strconv.Itoa(int(r.Result)) + separator + line
	}
	return strings.Join(lines, "\r\n")
}

func LastLine(r Reply) string {
	s := strings.Split(r.Message, "\n")
	return strings.TrimSpace(s[len(s)-1])
}

func (c *Client) SendCmd(cmd Command) {
	msg := cmd.Verb
	if len(cmd.Args) > 0 {
		msg += " " + cmd.Args
	}
	_, err := c.Rdr.WriteString(msg + "\r\n")
	c.Rdr.Flush()
	if err != nil {
		c.Conn.Close()
		c.State = Disconnected
	}
}

func (c *Client) Send(r Reply) {
	if r.Result == Ignore {
		return
	}
	_, err := c.Rdr.WriteString(FormatReply(r) + "\r\n")
	c.Rdr.Flush()
	if err != nil {
		c.Conn.Close()
		c.State = Disconnected
	}
}

func (c *Client) Receive() (Command, error) {
	str, err := c.Rdr.ReadString('\n')
	str = strings.TrimSpace(str)
	if err != nil {
		c.State = Disconnected
		return Command{}, err
	}
	split := strings.Split(str, " ")
	return Command{
		Verb: split[0],
		Args: strings.Join(split[1:], " "),
	}, nil
}
