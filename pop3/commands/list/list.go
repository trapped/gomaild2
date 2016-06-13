package list

import (
	"fmt"
	"github.com/trapped/gomaild2/db"
	. "github.com/trapped/gomaild2/pop3/structs"
	"strconv"
	"strings"
)

// Arguments:
// a message-number (optional), which, if present, may NOT
// refer to a message marked as deleted
func Process(c *Client, cmd Command) Reply {
	if c.State != Transaction {
		return Reply{
			Result:  ERR,
			Message: "invalid state",
		}
	}
	envs := db.List(c.GetString("authenticated_as"), false)
	if cmd.Args == "" {
		cnt, sz := db.Stat(c.GetString("authenticated_as"))
		msg := []string{fmt.Sprintf("%v messages (%v octets)", cnt, sz)}
		for i, env := range envs {
			msg = append(msg, fmt.Sprintf("%v %v", i, len(env.Body)))
		}
		msg = append(msg, ".")
		return Reply{
			Result:  OK,
			Message: strings.Join(msg, "\r\n"),
		}
	} else if cmd.Args != "" {
		for i, env := range envs {
			if strconv.Itoa(i) == cmd.Args {
				return Reply{
					Result:  OK,
					Message: fmt.Sprintf("%v %v", i, len(env.Body)),
				}
			}
		}
		return Reply{
			Result:  ERR,
			Message: "no such message",
		}
	} else {
		return Reply{
			Result:  ERR,
			Message: "syntax error in command arguments",
		}
	}
}
