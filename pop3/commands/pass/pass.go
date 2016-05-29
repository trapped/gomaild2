package pass

import (
	. "github.com/trapped/gomaild2/pop3/structs"
)

// Arguments:
// a server/mailbox-specific password (required)
// Restrictions:
// may only be given in the AUTHORIZATION state immediately
// after a successful USER command
func Process(c *Client, cmd Command) Reply {
	res := OK
	msg := ""
	if c.State != Authorization { //TODO USER must be last command issued by client as well
		res = ERR
		msg = "invalid state"
	}
	return Reply{Result: res, Message: msg}
}
