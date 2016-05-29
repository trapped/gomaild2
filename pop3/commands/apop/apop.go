package apop

import (
	. "github.com/trapped/gomaild2/pop3/structs"
)

// Arguments:
// a string identifying a mailbox and a MD5 digest string
// (both required)
// Restrictions:
// may only be given in the AUTHORIZATION state after the POP3
// greeting or after an unsuccessful USER or PASS command
func Process(c *Client, cmd Command) Reply {
	res := OK
	msg := ""
	if c.State != Authorization {
		res = ERR
		msg = "invalid state"
	}
	return Reply{Result: res, Message: msg}
}
