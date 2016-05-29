package capa

import (
	. "github.com/trapped/gomaild2/pop3/structs"
)

func Process(c *Client, cmd Command) Reply {
	return Reply{Result: OK, Message: "Capability list follows\r\nSASL PLAIN LOGIN CRAM-MD5\r\n."}
}
