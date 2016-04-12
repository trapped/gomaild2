package smtp

import (
	"strings"
	. "trapped/gomaild2/smtp/structs"

	"trapped/gomaild2/smtp/commands/quit"
)

func Process(client *Client, cmd Command) Reply {
	switch strings.ToLower(cmd.Verb) {

	case "quit":
		return quit.Process(client, cmd)

	default:
		return Reply{
			Result:  CommandNotImplemented,
			Message: "command not implemented",
		}
	}
}
