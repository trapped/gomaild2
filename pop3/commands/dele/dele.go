package dele

import (
	"github.com/trapped/gomaild2/db"
	. "github.com/trapped/gomaild2/pop3/structs"

	"fmt"
	"strconv"
)

// Arguments:
// a message-number (required) which may NOT refer to a
// message marked as deleted

//TODO add a config variable to only mark as deleted
func Process(c *Client, cmd Command) Reply {
	if c.State != Transaction {
		return Reply{Result: ERR, Message: "invalid state"}
	}
	if cmd.Args == "" {
		return Reply{Result: ERR, Message: "index required"}
	} else if cmd.Args != "" {
		envs := db.List(c.GetString("authenticated_as"), true)
		deleted := c.GetStringSlice("deleted")
		index, err := strconv.Atoi(cmd.Args)
		if err != nil {
			return Reply{Result: ERR, Message: "invalid argument"}
		}
		if index > len(envs) {
			return Reply{Result: ERR, Message: "no message found"}
		}
		for i, env := range envs {
			if strconv.Itoa(i) == cmd.Args {
				if c.IsDeleted(env.ID) {
					return Reply{Result: ERR, Message: fmt.Sprintf("message %v already deleted", i)}
				}
				deleted = append(deleted, env.ID)
				c.Set("deleted", deleted)
				env.Deleted = true
				env.Save()
				return Reply{Result: OK, Message: fmt.Sprintf("message %v deleted", i)}
			}
		}

	}
	return Reply{Result: ERR, Message: "invalid argument syntax"}
}
