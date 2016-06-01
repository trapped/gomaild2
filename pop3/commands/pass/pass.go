package pass

import (
	log "github.com/sirupsen/logrus"
	"github.com/trapped/gomaild2/db"
	"github.com/trapped/gomaild2/pop3/locker"
	. "github.com/trapped/gomaild2/pop3/structs"
	"strings"
)

// Arguments:
// a server/mailbox-specific password (required)
// Restrictions:
// may only be given in the AUTHORIZATION state immediately
// after a successful USER command
func Process(c *Client, cmd Command) Reply {
	if c.State != Authorization { //TODO USER must be last command issued by client as well
		return Reply{Result: ERR, Message: "invalid state"}
	}
	if c.GetBool("authenticated") {
		return Reply{Result: ERR, Message: "already authenticated"}
	}
	lastData := strings.Split(c.GetString("last_command"), ":")
	if len(lastData) != 2 {
		return Reply{Result: ERR, Message: "invalid arguments"}
	}
	lastCMD, user := lastData[0], lastData[1]
	if lastCMD != "USER" {
		return Reply{Result: ERR, Message: "illegal user and pass sequence"}
	}
	c.Set("last_command", "PASS")

	if pw, exists := db.Users()[user]; exists && pw == cmd.Args {
		c.Data = make(map[string]interface{})
		c.Set("authenticated", true)
		c.Set("authenticated_as", user)

		locker.Lock(user)
		c.State = Transaction

		log.WithFields(log.Fields{
			"id":   c.ID,
			"user": user,
		}).Info("Logged in")

		return Reply{
			Result:  OK,
			Message: "authentication successful",
		}
	}

	return Reply{Result: ERR, Message: "authentication failed"}
}
