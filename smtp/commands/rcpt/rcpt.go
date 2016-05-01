package rcpt

import (
	"errors"
	. "github.com/trapped/gomaild2/smtp/structs"
	"regexp"
	"strings"
)

var (
	rxExtractAddr *regexp.Regexp = regexp.MustCompile("^TO:<(.*)>$")
)

func parseTo(args string) (string, error) {
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
