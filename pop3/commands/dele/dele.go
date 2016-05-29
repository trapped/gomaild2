package dele

import (
	. "github.com/trapped/gomaild2/pop3/structs"
)

// Arguments:
// a message-number (required) which may NOT refer to a
// message marked as deleted

//TODO add a config variable to only mark as deleted
func Process(c *Client, cmd Command) Reply {
	res := OK
	msg := ""
	if c.State != Transaction {
		res = ERR
		msg = "invalid state"
	}
	return Reply{Result: res, Message: msg}
}
