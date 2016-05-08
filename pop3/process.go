package pop3

//WARNING: Automatically generated file. DO NOT EDIT!

import (
	. "github.com/mbags/gomaild2/pop3/structs"
	log "github.com/sirupsen/logrus"
	"strings"

	"github.com/mbags/gomaild2/pop3/commands/apop"

	"github.com/mbags/gomaild2/pop3/commands/auth"

	"github.com/mbags/gomaild2/pop3/commands/dele"

	"github.com/mbags/gomaild2/pop3/commands/list"

	"github.com/mbags/gomaild2/pop3/commands/noop"

	"github.com/mbags/gomaild2/pop3/commands/pass"

	"github.com/mbags/gomaild2/pop3/commands/quit"

	"github.com/mbags/gomaild2/pop3/commands/retr"

	"github.com/mbags/gomaild2/pop3/commands/rset"

	"github.com/mbags/gomaild2/pop3/commands/stat"

	"github.com/mbags/gomaild2/pop3/commands/stls"

	"github.com/mbags/gomaild2/pop3/commands/top"

	"github.com/mbags/gomaild2/pop3/commands/uidl"

	"github.com/mbags/gomaild2/pop3/commands/user"
)

func Process(c *Client, cmd Command) (reply Reply) {
	switch strings.ToLower(cmd.Verb) {

	case "apop":
		reply = apop.Process(c, cmd)

	case "auth":
		reply = auth.Process(c, cmd)

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
		//Result: CommandNotImplemented,
		//Message: "command not implemented",
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
		//"reply": LastLine(reply),
	}).Info("status")
	return
}
