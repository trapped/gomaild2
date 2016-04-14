package noop

import (
	. "github.com/trapped/gomaild2/smtp/structs"
)

func Process(c *Client, cmd Command) Reply {
	return Reply{
		Result:  Ok,
		Message: "no action performed",
	}
}
