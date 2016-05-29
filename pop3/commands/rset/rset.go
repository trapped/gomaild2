package rset

import (
	. "github.com/trapped/gomaild2/pop3/structs"
)

func Process(c *Client, cmd Command) Reply {
	res := OK
	msg := ""
	if c.State != Transaction {
		res = ERR
		msg = "invalid state"
	}
	return Reply{Result: res, Message: msg}
}
