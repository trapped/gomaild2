package rcpt

import (
	"errors"
	. "github.com/trapped/gomaild2/smtp/structs"
	"strings"
)

func parseTo(args string) (string, error) {
	split := strings.Split(args, ":")
	if len(split) != 2 {
		return "", errors.New("syntax error in command arguments")
	}
	for i, s := range split {
		split[i] = strings.TrimSpace(s)
	}
	if strings.ToUpper(split[0]) != "TO" {
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

func appendRecipient(c *Client, new_recipient string) {
	recipients := c.Data["recipients"]
	if recipients == nil {
		recipients = interface{}(make([]string, 0))
	}
	recipients = append(recipients.([]string), new_recipient)
	c.Data["recipients"] = recipients
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
	appendRecipient(c, recipient)
	c.State = ReceivingHeaders
	return Reply{
		Result:  Ok,
		Message: "recipient added",
	}
}
