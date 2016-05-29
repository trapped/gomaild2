package pop3

import (
	"bufio"
	"net"
	"strings"
	"time"
)

const (
	OK   string = "+OK"
	ERR  string = "-ERR"
	TERM string = "."
)

const (
	SessionTimeout = time.Minute * 10
)

type SessionState int

var Extensions []string = []string{
	"PIPELINING",
}

const (
	Authorization SessionState = iota
	Transaction
	Update
	Disconnected
)

type Command struct {
	Verb string
	Args string
}

type Reply struct {
	Result  string
	Message string
}

type Client struct {
	Conn  net.Conn
	ID    string
	State SessionState
	Data  map[string]interface{}
	Rdr   *bufio.ReadWriter
}

func (c *Client) MakeReader() {
	c.Rdr = bufio.NewReadWriter(bufio.NewReader(c.Conn), bufio.NewWriter(c.Conn))
}

func (c *Client) Send(r Reply) {
	if r.Result == "" {
		return
	}
	_, err := c.Rdr.WriteString(r.Result + " " + r.Message + "\r\n")
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
