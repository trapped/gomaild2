package stat

import (
	"fmt"
	"github.com/trapped/gomaild2/db"
	. "github.com/trapped/gomaild2/pop3/structs"
)

func Process(c *Client, cmd Command) Reply {
	if c.State != Transaction {
		return Reply{
			Result:  ERR,
			Message: "invalid state",
		}
	}

	cnt, sz := db.Stat(c.GetString("authenticated_as"))

	return Reply{
		Result:  OK,
		Message: fmt.Sprintf("%v %v", cnt, sz),
	} // want : +OK XX(#emails) YYY(#bytes)
}
