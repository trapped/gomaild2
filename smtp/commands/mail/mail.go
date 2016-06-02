package mail

import (
	"errors"
	. "github.com/trapped/gomaild2/smtp/structs"
	"regexp"
)

var (
	rxExtractAddr *regexp.Regexp = regexp.MustCompile("(?i)^FROM:( )*<(.*)>(.*)$")
)

func parseFrom(args string) (string, error) {
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
	//TODO: abstract require_auth in process.go.tmpl
	if c.GetBool("require_auth") && !c.GetBool("authenticated") {
		return Reply{
			Result:  AuthenticationRequired,
			Message: "authentication required",
		}
	}
	if c.State != Identified {
		return Reply{
			Result:  BadSequence,
			Message: "wrong command sequence",
		}
	}
	sender, err := parseFrom(cmd.Args)
	if err != nil {
		return Reply{
			Result:  ArgumentSyntaxError,
			Message: err.Error(),
		}
	}
	//TODO: check blacklist
	c.Set("sender", sender)
	c.State = ReceivingHeaders
	return Reply{
		Result:  Ok,
		Message: "sender validated",
	}
}
