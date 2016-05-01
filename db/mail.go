package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	"net/mail"
	"time"
)

type EnvelopeType uint8
type EnvelopeStatus uint8

const (
	Inbound   EnvelopeType   = 0
	Outbound                 = 1
	Delivered                = 0
	Pending   EnvelopeStatus = 1 //only meaningful if Type is Outbound
	Failed                   = 2 //DeliverTries exceeded maximum tries
)

type Envelope struct {
	Type            EnvelopeType
	Status          EnvelopeStatus //checked by outbound delivery agent during dispatch
	DeliverTries    int            //times the delivery agent has tried delivering the envelope
	NextDeliverTime time.Time      //checked periodically by outbound delivery agent
	Session         string         //client session id
	OutboundAllowed bool           //whether envelope is allowed to be delivered to external recipients
	Sender          string
	Recipients      []string
	BodySize        int
	Date            time.Time
}

/*
  user:
    inbound:  json(Envelope) = body
    outbound: json(Envelope) = body
*/

func (e *Envelope) Save(body string) error {
	return db.Update(func(tx *bolt.Tx) error {
		users := Users()
		ext_recipients := make([]string, 0)
		_, sender_local := users[e.Sender]

		env_json, err := json.Marshal(e)
		if err != nil {
			log.Fatal(err)
		}

		env_log := log.WithFields(log.Fields{
			"session":        e.Session,
			"sender":         e.Sender,
			"sender_local":   sender_local,
			"recipients":     e.Recipients,
			"body_size":      e.BodySize,
			"ext_recipients": ext_recipients,
		})

		message_id := ""
		return_path := ""
		msg, err := mail.ReadMessage(bytes.NewReader([]byte(body)))
		if err != nil {
			env_log.Error(err)
		} else {
			if msg.Header.Get("Return-Path") == "" {
				return_path = fmt.Sprintf("Return-Path: <%v>\r\n", e.Sender)
			}
			if msg.Header.Get("Message-ID") == "" {
				return_path = fmt.Sprintf("Message-ID: <%v.%v@%v>\r\n", string(MD5(env_json)), string(MD5([]byte(body))), config.GetString("server.name"))
			}
		}
		headers := message_id + return_path

		for _, rcpt := range e.Recipients {
			_, rcpt_local := users[rcpt]

			//if recipient is local, add envelope to their inbox
			if rcpt_local || config.GetBool("db.save_all_mail") {
				delivered_to := fmt.Sprintf("Delivered-To: %v\r\n", rcpt)
				r_headers := headers + delivered_to

				b, err := userBucket(tx, rcpt)
				inbound := b.Bucket([]byte("inbound"))

				err = inbound.Put(env_json, []byte(r_headers+body))
				if err != nil {
					env_log.Error(err)
					return fmt.Errorf("failed to queue")
				}
				env_log.Info("Saved envelope to DB")
			}

			//if recipient is not local but sender is, flag recipient as external
			if !rcpt_local && sender_local {
				ext_recipients = append(ext_recipients, rcpt)
			}
		}

		//if sender is local and there are external recipients, add envelope to sender outbox
		if sender_local && len(ext_recipients) > 0 {
			if !e.OutboundAllowed {
				env_log.Error("Blocked outbound mail")
				return fmt.Errorf("outbound mail not allowed")
			}

			b, err := userBucket(tx, e.Sender)
			if err != nil {
				env_log.Error("Couldn't access sender box: ", err)
				return fmt.Errorf("couldn't access sender box")
			}

			//convert envelope to Outbound
			o_env := &Envelope{}
			*o_env = *e
			o_env.Type = Outbound
			o_env.Status = Pending

			env_json, err := json.Marshal(o_env)
			if err != nil {
				log.Fatal(err)
			}

			outbound := b.Bucket([]byte("outbound"))
			err = outbound.Put(env_json, []byte(body))
			if err != nil {
				env_log.Error("Couldn't access sender box: ", err)
				return fmt.Errorf("couldn't access sender box")
			}
		}

		return nil
	})
}
