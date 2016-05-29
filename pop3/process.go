package pop3

//WARNING: Automatically generated file. DO NOT EDIT!

import (
	log "github.com/sirupsen/logrus"
	. "github.com/trapped/gomaild2/pop3/structs"
	"strings"

	"github.com/trapped/gomaild2/pop3/commands/apop"

	"github.com/trapped/gomaild2/pop3/commands/auth"

	"github.com/trapped/gomaild2/pop3/commands/capa"

	"github.com/trapped/gomaild2/pop3/commands/dele"

	"github.com/trapped/gomaild2/pop3/commands/list"

	"github.com/trapped/gomaild2/pop3/commands/noop"

	"github.com/trapped/gomaild2/pop3/commands/pass"

	"github.com/trapped/gomaild2/pop3/commands/quit"

	"github.com/trapped/gomaild2/pop3/commands/retr"

	"github.com/trapped/gomaild2/pop3/commands/rset"

	"github.com/trapped/gomaild2/pop3/commands/stat"

	"github.com/trapped/gomaild2/pop3/commands/stls"

	"github.com/trapped/gomaild2/pop3/commands/top"

	"github.com/trapped/gomaild2/pop3/commands/uidl"

	"github.com/trapped/gomaild2/pop3/commands/user"
)

func Process(c *Client, cmd Command) (reply Reply) {
	switch strings.ToLower(cmd.Verb) {

	case "apop":
		reply = apop.Process(c, cmd)

	case "auth":
		reply = auth.Process(c, cmd)

	case "capa":
		reply = capa.Process(c, cmd)

	case "dele":
		reply = dele.Process(c, cmd)

	case "list":
		reply = list.Process(c, cmd)

	case "noop":
		reply = noop.Process(c, cmd)

	case "pass":
		reply = pass.Process(c, cmd)

	case "quit":
		reply = quit.Process(c, cmd)

	case "retr":
		reply = retr.Process(c, cmd)

	case "rset":
		reply = rset.Process(c, cmd)

	case "stat":
		reply = stat.Process(c, cmd)

	case "stls":
		reply = stls.Process(c, cmd)

	case "top":
		reply = top.Process(c, cmd)

	case "uidl":
		reply = uidl.Process(c, cmd)

	case "user":
		reply = user.Process(c, cmd)

	default:
		reply = Reply{
			Result:  ERR,
			Message: "command not implemented",
		}
	}
	//if reply.Result == Ignore {
	//return
	//}
	log.WithFields(log.Fields{
		"id":     c.ID,
		"cmd":    cmd.Verb,
		"args":   cmd.Args,
		"result": reply.Result,
	}).Info("status")
	return
}
