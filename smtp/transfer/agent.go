package transfer

import (
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	. "github.com/trapped/gomaild2/db"
	. "github.com/trapped/gomaild2/structs"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type Agent struct {
}

var pipeline chan *Envelope = make(chan *Envelope)

//TODO: send warnings to sender
func deliver(host string, c *smtp.Client, env *Envelope, recipients []string) error {
	defer func() {
		c.Quit()
		c.Close()
	}()
	//EHLO
	err := c.Hello(config.GetString("server.name"))
	if err != nil {
		return err
	}
	//STARTTLS if supported
	supports_tls, _ := c.Extension("STARTTLS")
	if !supports_tls && !config.GetBool("transfer.allow_unencrypted") {
		return fmt.Errorf("Host doesn't support TLS and unencrypted delivery is disabled")
	}
	if supports_tls {
		conf := &tls.Config{
			InsecureSkipVerify: config.GetBool("transfer.allow_insecure"),
			ServerName:         host,
		}
		err = c.StartTLS(conf)
		if err != nil && !config.GetBool("transfer.allow_unencrypted") {
			return err
		}
	}
	//MAIL FROM
	err = c.Mail(env.Sender)
	if err != nil {
		return err
	}
	//RCPT TO
	failed_recipients := make([]string, 0)
	for _, recipient := range recipients {
		err = c.Rcpt(recipient)
		if err != nil {
			failed_recipients = append(failed_recipients, recipient)
		}
	}
	if len(failed_recipients) == len(recipients) {
		return fmt.Errorf("Host rejected all recipients")
	} else if len(failed_recipients) < len(recipients) {
		//send warning to sender
	}
	//DATA
	io, err := c.Data()
	if err != nil {
		return err
	}
	_, err = io.Write([]byte(env.Body))
	if err != nil {
		return err
	}
	//QUIT
	return nil
}

//builds a map of domains and their respective users from the recipients
func aggregateRecipients(recipients []string) map[string][]string {
	domains := make(map[string][]string, 0)
	for _, recipient := range recipients {
		domain := strings.Split(recipient, "@")[1]
		if rcpts, ok := domains[domain]; ok {
			domains[domain] = append(rcpts, recipient)
		} else {
			rcpts := make([]string, 0)
			domains[domain] = append(rcpts, recipient)
		}
	}
	return domains
}

//TODO: send warnings to sender
func worker() {
	for {
		env := <-pipeline
		log_e := log.WithFields(log.Fields{
			"env":        env.ID,
			"sender":     env.Sender,
			"recipients": env.Recipients,
		})
		for domain, recipients := range aggregateRecipients(env.Recipients) {
			log_ed := log_e.WithFields(log.Fields{
				"domain":     domain,
				"recipients": recipients,
			})
			mxs, err := net.LookupMX(domain)
			if err != nil {
				log_ed.WithField("err", err).Error("Couldn't lookup MX")
				//send warning to sender
				continue
			}
			delivered := false
			for i := 0; i < len(mxs); i++ {
				host := mxs[i].Host
				log_edh := log_ed.WithField("host", host)
				var client *smtp.Client
				var tries int
				for tries = 0; tries < config.GetInt("transfer.max_tries"); tries++ {
					client, err = smtp.Dial(host + ":25")
					if err != nil {
						log_edh.WithField("err", err).Error("Connection failed")
						client = nil
						//couldn't connect to this host, try again
						continue
					}
					break
				}
				if client == nil {
					//couldn't connect to this host, try next one
					log_edh.Error("Max connection tries reached")
					continue
				}
				err := deliver(host, client, env, recipients)
				if err != nil {
					log_edh.WithField("err", err).Error("Delivery failed for this host")
					//couldn't deliver to this host, try another one
					continue
				} else {
					delivered = true
					break
				}
			}
			if !delivered {
				env.Status = Failed
				env.Save()
				log_ed.Error("Delivery failed for this domain")
				//TODO: set NextDeliverTime
				//if transfer.tenacious == true, schedule for another try at a later time
				//and send "temporary failure" warning to sender
				//otherwise, send "permanent failure" warning to sender
			} else {
				//mark envelope as delivered
				env.Status = Delivered
				env.Save()
				log_e.Info("Delivery successful")
			}
		}
	}
}

func (a *Agent) Start() {
	WaitConfig("config.loaded")
	for i := 0; i < config.GetInt("transfer.worker_count"); i++ {
		go worker()
	}
	for {
		mail := Sweep()
		for _, env := range mail {
			log.WithField("env", env.ID).Info("Sweeping")
			pipeline <- env
		}
		time.Sleep(time.Minute)
	}
}
