package top

import (
	. "github.com/trapped/gomaild2/pop3/structs"
)

// Arguments:
// a message-number (required) which may NOT refer to to a
// message marked as deleted, and a non-negative number
// of lines (required)
// Restrictions:
// may only be given in the TRANSACTION state
func Process(c *Client, cmd Command) Reply {
	res := OK
	msg := ""
	if c.State != Transaction {
		res = ERR
		msg = "invalid state"
	}
	return Reply{Result: res, Message: msg}
}
