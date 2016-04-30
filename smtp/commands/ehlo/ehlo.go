package ehlo

import (
	"github.com/trapped/gomaild2/smtp/commands/helo"
	. "github.com/trapped/gomaild2/smtp/structs"
	"strings"
)

func Process(c *Client, cmd Command) Reply {
	if c.State >= Identified {
		return helo.AlreadyIdentified(c, cmd)
	} else {
		reply := helo.Identify(c, cmd)
		if IsSuccess(reply) {
			reply.Message = strings.Join(append(Extensions, reply.Message), "\n")
		}
		return reply
	}
}
