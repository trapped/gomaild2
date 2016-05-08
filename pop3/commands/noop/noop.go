package noop

import (
	. "github.com/mbags/gomaild2/pop3/structs"
)

func Process(c *Client, cmd Command) Reply {
	res := ""
	if c.State == Transaction {
		res = OK
	}
	return Reply{
		Result:  res,
		Message: "no action performed",
	}
}
