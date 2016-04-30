package mail

import (
	"errors"
	. "github.com/trapped/gomaild2/smtp/structs"
	"strings"
)

func parseFrom(args string) (string, error) {
	split := strings.Split(args, ":")
	if len(split) != 2 {
		return "", errors.New("syntax error in command arguments")
	}
	for i, s := range split {
		split[i] = strings.TrimSpace(s)
	}
	if strings.ToUpper(split[0]) != "FROM" {
		return "", errors.New("syntax error in command arguments")
	}
	enc_addr := split[1]
	if !strings.HasPrefix(enc_addr, "<") || !strings.HasSuffix(enc_addr, ">") {
		return "", errors.New("syntax error in command arguments")
	}
	addr := strings.TrimSpace(enc_addr[1 : len(enc_addr)-1])
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
