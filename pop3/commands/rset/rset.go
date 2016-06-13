package rset

import (
	"fmt"

	"github.com/trapped/gomaild2/db"
	. "github.com/trapped/gomaild2/pop3/structs"
)

func Process(c *Client, cmd Command) Reply {
	if c.State != Transaction {
		return Reply{Result: ERR, Message: "invalid state"}
	}
	envs := db.List(c.GetString("authenticated_as"), true)
	for _, env := range envs {
		if c.IsDeleted(env.ID) {
			env.Deleted = false
			env.Save()
		}
	}
	c.Set("deleted", []string{})
	cnt, sz := db.Stat(c.GetString("authenticated_as"))

	return Reply{
		Result:  OK,
		Message: fmt.Sprintf("%v %v", cnt, sz),
	}
}
