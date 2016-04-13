package smtp

//WARNING: Automatically generated file. DO NOT EDIT!

import (
	. "github.com/trapped/gomaild2/smtp/structs"
	"strings"

	"github.com/trapped/gomaild2/smtp/commands/helo"

	"github.com/trapped/gomaild2/smtp/commands/mail"

	"github.com/trapped/gomaild2/smtp/commands/quit"

	"github.com/trapped/gomaild2/smtp/commands/rcpt"
)

func Process(client *Client, cmd Command) Reply {
	switch strings.ToLower(cmd.Verb) {

	case "helo":
		return helo.Process(client, cmd)

	case "mail":
		return mail.Process(client, cmd)

	case "quit":
		return quit.Process(client, cmd)

	case "rcpt":
		return rcpt.Process(client, cmd)

	default:
		return Reply{
			Result:  CommandNotImplemented,
			Message: "command not implemented",
		}
	}
}
