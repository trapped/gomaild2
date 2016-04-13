package data

import (
	"fmt"
	. "github.com/trapped/gomaild2/smtp/structs"
	"strings"
)

func queueMessage(sender string, recipients []string, body string) {
	fmt.Printf("Sender: %v\nRecipients: %v\nBody:\n\t%v\n", sender,
		strings.Join(recipients, ", "), strings.Join(strings.Split(body, "\n"), "\n\t"))
}

func Process(c *Client, cmd Command) Reply {
	if c.State != ReceivingHeaders && len(c.Data["recipients"].([]string)) < 0 {
		return Reply{
			Result:  BadSequence,
			Message: "wrong command sequence",
		}
	}
	c.State = ReceivingBody
	c.Send(Reply{
		Result:  StartSending,
		Message: "start sending input",
	})
	body := ""
	for {
		line, err := c.Rdr.ReadString('\n')
		if err != nil {
			c.State = Disconnected
			return Reply{}
		}
		if line == ".\r\n" && strings.HasSuffix(body, "\r\n") {
			queueMessage(c.Data["sender"].(string),
				c.Data["recipients"].([]string), body)
			c.Data["sender"] = nil
			c.Data["recipients"] = nil
			return Reply{
				Result:  Ok,
				Message: "queued",
			}
		} else {
			body += line
		}
	}
}
