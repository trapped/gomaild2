package user

import (
	"github.com/trapped/gomaild2/db"
	. "github.com/trapped/gomaild2/pop3/structs"
)

// Arguments:
// a string identifying a mailbox (required), which is of
// significance ONLY to the server
// Restrictions:
// may only be given in the AUTHORIZATION state after the POP3
// greeting or after an unsuccessful USER or PASS command
func Process(c *Client, cmd Command) Reply {

	if c.State != Authorization {
		return Reply{Result: ERR, Message: "invalid state"}
	}
	if c.GetBool("authenticated") {
		return Reply{Result: ERR, Message: "already authenticated"}
	}
	user := cmd.Args
	if _, exists := db.Users()[user]; exists {
		c.Set("last_command", "USER:"+user)
		return Reply{Result: OK, Message: user + " is a valid mailbox"}
	}

	return Reply{Result: ERR, Message: "sorry couldn't aquire mailbox for " + user}
}
