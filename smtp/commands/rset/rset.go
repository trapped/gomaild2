package rset

import (
	. "github.com/trapped/gomaild2/smtp/structs"
)

func Process(c *Client, cmd Command) Reply {
	delete(c.Data, "sender")
	delete(c.Data, "recipients")
	c.State = Identified
	return Reply{
		Result:  Ok,
		Message: "reset envelope",
	}
}
