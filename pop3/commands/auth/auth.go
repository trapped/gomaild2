package auth

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	config "github.com/spf13/viper"
	"github.com/trapped/gomaild2/db"
	. "github.com/trapped/gomaild2/pop3/structs"
	. "github.com/trapped/gomaild2/structs"
	"strings"
)

var (
	usernameString string = base64.StdEncoding.EncodeToString([]byte("Username:"))
	passwordString string = base64.StdEncoding.EncodeToString([]byte("Password:"))
)

func init() {
	Extensions = append(Extensions, "AUTH PLAIN LOGIN CRAM-MD5")
}
func decodeCreds(auth string) (string, string, string) {
	decoded, _ := base64.StdEncoding.DecodeString(auth)
	split := strings.Split(string(decoded), "\x00")
	if len(split) == 3 {
		return split[0], split[1], split[2]
	}
	return "", "", ""
}

func verify(c *Client, username string, password string) Reply {
	if pw, exists := db.Users()[username]; exists && pw == password {
		c.Set("authenticated", true)
		c.Set("authenticated_as", username)
		//TODO: aquire lock here on db
		c.State = Transaction
		return Reply{
			Result:  OK,
			Message: "authentication successful",
		}
	}
	return Reply{
		Result:  ERR,
		Message: "authentication failed",
	}
}

func processAuth(c *Client, method string, auth string) Reply {
	user, pass := "", ""
	switch strings.ToUpper(method) {
	case "PLAIN": // Ex: XXX\x00example@test.com\x00testpassword = eHh4eABleGFtcGxlQHRlc3QuY29tAHRlc3RwYXNzd29yZA==
		if auth == "" {
			authcmd, _ := c.Receive()
			auth = authcmd.Verb
		}
		_, user, pass = decodeCreds(auth) // some clients ditch zid anyway
	case "LOGIN":
		if auth == "" {
			c.Send(Reply{Result: OK, Message: usernameString})
			authcmd, _ := c.Receive()
			auth = authcmd.Verb
		}
		c.Send(Reply{Result: OK, Message: passwordString})
		auth_pw, _ := c.Receive()
		auth = auth + "\x00" + auth + "\x00" + auth_pw.Verb
		username_b, _ := base64.StdEncoding.DecodeString(auth)
		password_b, _ := base64.StdEncoding.DecodeString(auth_pw.Verb)
		user, pass = string(username_b), string(password_b)
	case "CRAM-MD5":
		challenge := fmt.Sprintf("<%v.%v.%v@%v>", c.ID, SessionID(6), SessionID(6), config.GetString("server.name"))
		c.Send(Reply{Result: OK, Message: base64.StdEncoding.EncodeToString([]byte(challenge))})
		auth_c, _ := c.Receive()
		digest, _ := base64.StdEncoding.DecodeString(auth_c.Verb)
		digest_split := strings.Split(string(digest), " ")
		user = digest_split[0]
		pw, exists := db.Users()[user]
		if exists {
			mac := hmac.New(md5.New, []byte(pw))
			mac.Write([]byte(challenge))
			real_digest := hex.EncodeToString(mac.Sum(nil))
			if digest_split[1] == real_digest {
				pass = pw
			}
		}
	default:
		return Reply{Result: ERR, Message: "authentication strategy not supported"}
	}
	return verify(c, user, pass)
}

func Process(c *Client, cmd Command) Reply {
	if c.State != Authorization {
		return Reply{Result: ERR, Message: "bad state"}
	}
	if c.GetBool("authenticated") {
		return Reply{Result: ERR, Message: "already authenticated"}
	}

	args := strings.Split(cmd.Args, " ")

	if len(args) == 1 {
		return processAuth(c, args[0], "")
	} else if len(args) == 2 {
		return processAuth(c, args[0], args[1])
	} else {
		return Reply{
			Result:  ERR,
			Message: "unrecognized authentication method",
		}
	}
}
