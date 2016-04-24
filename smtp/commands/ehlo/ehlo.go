package ehlo

import (
	"github.com/trapped/gomaild2/smtp/commands/helo"
	. "github.com/trapped/gomaild2/smtp/structs"
	"strings"
)

func GetExtensions() []string {
	//TODO: actually check supported extensions
	return []string{"PIPELINING", "8BITMIME"}
}

func Process(c *Client, cmd Command) Reply {
	if c.State >= Identified {
		return helo.AlreadyIdentified(c, cmd)
	} else {
		reply := helo.Identify(c, cmd)
		if IsSuccess(reply) {
			reply.Message = strings.Join(append(GetExtensions(), reply.Message), "\n")
		}
		return reply
	}
}
