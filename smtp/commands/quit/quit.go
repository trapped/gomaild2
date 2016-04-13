package quit

import (
	. "github.com/trapped/gomaild2/smtp/structs"
)

func Process(c *Client, cmd Command) Reply {
	c.State = Disconnected
	return Reply{
		Result:  Closing,
		Message: "see you next time",
	}
}
