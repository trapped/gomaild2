package helo

import (
	"regexp"
	. "trapped/gomaild2/smtp/structs"
)

var domain_regex *regexp.Regexp = regexp.MustCompile("^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])$")
var ip_regex *regexp.Regexp = regexp.MustCompile("^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$")

func identify(c *Client, cmd Command) Reply {
	valid := domain_regex.MatchString(cmd.Args) || ip_regex.MatchString(cmd.Args)
	//TODO: start blacklist check
	if valid {
		c.Data["identifier"] = interface{}(cmd.Args)
		c.State = Identified
		return Reply{
			Result:  Ok,
			Message: "domain validated",
		}
	} else {
		return Reply{
			Result:  ArgumentSyntaxError,
			Message: "invalid domain",
		}
	}
}

func alreadyidentified(c *Client, cmd Command) Reply {
	return Reply{
		Result:  AlreadyIdentified,
		Message: "already identified",
	}
}

func Process(c *Client, cmd Command) Reply {
	if c.State >= Identified {
		return alreadyidentified(c, cmd)
	} else {
		return identify(c, cmd)
	}
}
