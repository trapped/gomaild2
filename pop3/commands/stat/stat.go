package stat

import (
	"github.com/trapped/gomaild2/db"
	. "github.com/trapped/gomaild2/pop3/structs"
	"strconv"
)

func Process(c *Client, cmd Command) Reply {
	res := OK
	msg := ""
	if c.State != Transaction {
		res = ERR
		msg = "invalid state"
	}

	cnt, sz, err := db.Stat(c.GetString("authenticated_as"))
	if err != nil {
		msg = "couldn't get mailbox"
		return Reply{Result: ERR, Message: msg}
	}

	msg = strconv.FormatInt(int64(cnt), 10) + " " + strconv.FormatInt(int64(sz), 10)
	return Reply{Result: res, Message: msg} // want : +OK XX(#emails) YYY(#bytes)
}
