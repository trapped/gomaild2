package rcpt

import (
	"errors"
	. "github.com/trapped/gomaild2/smtp/structs"
	"regexp"
)

var (
	rxExtractAddr *regexp.Regexp = regexp.MustCompile("(?i)^TO:( )*<(.*)>(.*)$")
)

func parseTo(args string) (string, error) {
	if !rxExtractAddr.MatchString(args) {
		return "", errors.New("syntax error in command arguments")
	}

	addr := rxExtractAddr.FindStringSubmatch(args)[2]
	if !RxEmailAddress.MatchString(addr) {
		return "", errors.New("invalid address")
	}
	return addr, nil
}

func Process(c *Client, cmd Command) Reply {
	if c.State != ReceivingHeaders {
		return Reply{
			Result:  BadSequence,
			Message: "wrong command sequence",
		}
	}
	recipient, err := parseTo(cmd.Args)
	if err != nil {
		return Reply{
			Result:  ArgumentSyntaxError,
			Message: err.Error(),
		}
	}
	//DO NOT CHECK IF RECIPIENTS EXISTS (that'd leak data)
	c.Set("recipients", append(c.GetStringSlice("recipients"), recipient))
	c.State = ReceivingHeaders
	return Reply{
		Result:  Ok,
		Message: "recipient added",
	}
}
