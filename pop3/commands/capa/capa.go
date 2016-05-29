package capa

import (
	. "github.com/trapped/gomaild2/pop3/structs"
	"strings"
)

func Process(c *Client, cmd Command) Reply {
	return Reply{
		Result:  OK,
		Message: "capability list follows\r\n" + strings.Join(Extensions, "\r\n") + "\r\n.",
	}
}
