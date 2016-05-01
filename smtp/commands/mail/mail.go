package mail

import (
	"errors"
	. "github.com/trapped/gomaild2/smtp/structs"
	"regexp"
	"strings"
)

var (
	rxExtractAddr *regexp.Regexp = regexp.MustCompile("^FROM:<(.*)>$")
)

func parseFrom(args string) (string, error) {
	split := strings.Split(args, " ")
	if len(split) < 1 {
		return "", errors.New("syntax error in command arguments")
	}
	//cleanup args
	for i := 0; i < len(split); i++ {
		split[i] = strings.TrimSpace(split[i])
		if split[i] == "" {
			split = append(split[:i], split[i+1:]...)
			i--
		}
	}

	if !rxExtractAddr.MatchString(split[0]) {
		return "", errors.New("syntax error in command arguments")
	}

	addr := rxExtractAddr.FindStringSubmatch(split[0])[1]
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
