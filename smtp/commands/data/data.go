package data

import (
	"fmt"
	config "github.com/spf13/viper"
	"github.com/trapped/gomaild2/db"
	. "github.com/trapped/gomaild2/smtp/structs"
	. "github.com/trapped/gomaild2/structs"
	"strings"
	"time"
)

func receivedString(c *Client, e *db.Envelope) string {
	secure, authenticated := "", ""
	if c.GetBool("secure") {
		secure = "S"
	}
	if c.GetBool("authenticated") {
		authenticated = "A"
	}
	return fmt.Sprintf("Received: from %v (%v)\r\n\tby %v with ESMTP%v%v id %v;\r\n\t%v\r\n",
		c.GetString("identifier"), c.Conn.RemoteAddr().String(), config.GetString("server.name"),
		secure, authenticated, c.ID, e.Date.Format(time.RFC1123Z))
}

func queueMessage(c *Client, sender string, recipients []string, body string) error {
	envelope := &db.Envelope{
		Sender:          sender,
		Recipients:      recipients,
		Session:         c.ID,
		ID:              c.ID + "." + SessionID(12),
		OutboundAllowed: c.GetBool("outbound"),
		Date:            time.Now(),
		NextDeliverTime: time.Now(),
	}

	//TODO: check sender's domain SPF

	//TODO: add Received-SPF, Authentication-Results
	headers := ""
	headers += receivedString(c, envelope)
	envelope.Body = headers + body

	return envelope.Save()
}

func Process(c *Client, cmd Command) Reply {
	recipients := c.GetStringSlice("recipients")
	sender := c.GetString("sender")
	if c.State != ReceivingHeaders && len(recipients) < 0 {
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
			err = queueMessage(c, sender, recipients, body)
			c.Set("sender", nil)
			c.Set("recipients", nil)
			c.State = Identified
			if err != nil {
				return Reply{
					Result:  LocalError,
					Message: err.Error(),
				}
			} else {
				return Reply{
					Result:  Ok,
					Message: "queued",
				}
			}
		} else {
			body += line
		}
	}
}
