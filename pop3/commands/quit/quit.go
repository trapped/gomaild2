package quit

import (
	. "github.com/trapped/gomaild2/pop3/structs"
)

func Process(c *Client, cmd Command) Reply {
	res, msg := "", ""
	switch c.State {
	case Authorization:
		res = OK
		msg = "server signing off"
		c.State = Disconnected
	case Transaction:
		//TODO real transaction logic
		c.State = Update
		res = OK
		msg = "server signing off (entering update state)"
	}

	return Reply{
		Result:  res,
		Message: msg,
	}
}
