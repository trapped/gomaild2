package retr

import (
	"fmt"
	"github.com/trapped/gomaild2/db"
	. "github.com/trapped/gomaild2/pop3/structs"
	"strconv"
)

// Arguments:
// a message-number (required) which may NOT refer to a
// message marked as deleted
func Process(c *Client, cmd Command) Reply {
	if c.State != Transaction {
		return Reply{
			Result:  ERR,
			Message: "invalid state",
		}
	}
	envs := db.List(c.GetString("authenticated_as"), false)
	if cmd.Args != "" {
		for i, env := range envs {
			if strconv.Itoa(i) == cmd.Args {
				return Reply{
					Result:  OK,
					Message: fmt.Sprintf("%v octets\r\n", len(env.Body)) + env.Body + "\r\n.",
				}
			}
		}
	}
	return Reply{
		Result:  ERR,
		Message: "no such message",
	}
}
