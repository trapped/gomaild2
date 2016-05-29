package noop

import (
	. "github.com/trapped/gomaild2/pop3/structs"
)

func Process(c *Client, cmd Command) Reply {
	res := OK
	if c.State != Transaction {
		res = ERR
	}
	return Reply{
		Result:  res,
		Message: "no action performed",
	}
}
