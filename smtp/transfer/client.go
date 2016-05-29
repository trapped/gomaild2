package transfer

import (
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	//. "github.com/trapped/gomaild2/smtp/structs"
	"net"
)

type Client struct {
	Conn net.Conn
}

func (c *Client) Connect(addr string) (err error) {
	tries := 0
	log := log.WithFields(log.Fields{
		"remote_addr": addr,
		"tries":       tries,
	})
	for tries < config.GetInt("transfer.max_tries") {
		log.Info("Connecting")
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			tries++
		} else {
			c.Conn = conn
			log.Info("Connected")
			return nil
		}
	}
	return
}
