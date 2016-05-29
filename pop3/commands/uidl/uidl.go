package uidl

import (
	. "github.com/trapped/gomaild2/pop3/structs"
)

// Arguments:
// a message-number (optional), which, if present, may NOT
// refer to a message marked as deleted
// Restrictions:
// may only be given in the TRANSACTION state.
func Process(c *Client, cmd Command) Reply {
	res := OK
	msg := ""
	if c.State != Transaction {
		res = ERR
		msg = "invalid state"
	}
	return Reply{Result: res, Message: msg}
}
