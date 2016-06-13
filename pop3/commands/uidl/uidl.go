package uidl

import (
	"fmt"
	"github.com/trapped/gomaild2/db"
	. "github.com/trapped/gomaild2/pop3/structs"
	"strconv"
	"strings"
)

func init() {
	Extensions = append(Extensions, "UIDL")
}

// Arguments:
// a message-number (optional), which, if present, may NOT
// refer to a message marked as deleted
// Restrictions:
// may only be given in the TRANSACTION state.
func Process(c *Client, cmd Command) Reply {
	if c.State != Transaction {
		return Reply{
			Result:  ERR,
			Message: "invalid state",
		}
	}
	envs := db.List(c.GetString("authenticated_as"), false)
	if cmd.Args == "" { // clawsmail says this an invalid response
		msg := []string{"uid listing follows"}
		for i, env := range envs {
			msg = append(msg, fmt.Sprintf("%v %v", i, env.ID)) // shouldnt we count from 1 ?
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
					Message: fmt.Sprintf("%v %v", i, env.ID),
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
