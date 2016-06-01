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
	Pending   EnvelopeStatus = 0 //only meaningful if Type is Outbound
	Delivered                = 1
	Failed                   = 2 //DeliverTries exceeded maximum tries
	Assigned                 = 3
)

type Envelope struct {
	Type            EnvelopeType
	Status          EnvelopeStatus //checked by outbound delivery agent during dispatch
	DeliverTries    int            //times the delivery agent has tried delivering the envelope
	NextDeliverTime time.Time      //checked periodically by outbound delivery agent
	ID              string         //<client session id>:<random string>
	Session         string
	OutboundAllowed bool //whether envelope is allowed to be delivered to external recipients
	Sender          string
	Recipients      []string
	Body            string
	Date            time.Time
	Deleted         bool
}

func UnassignAll(tx *bolt.Tx) {
	for user, _ := range Users() {
		b, err := UserBucket(tx, user)
		if err != nil {
			log.Error(err)
		}
		bo := b.Bucket([]byte("outbound"))
		bo.ForEach(func(k, v []byte) error {
			env := &Envelope{}
			err := json.Unmarshal(v, env)
			if err != nil {
				log.Error(err)
			}
			if env.Status == Assigned {
				env.Status = Pending
				env_json, err := json.Marshal(env)
				if err != nil {
					log.Error(err)
				}
				bo.Put([]byte(env.ID), []byte(env_json))
			}
			return nil
		})
	}
}

func (e *Envelope) Save() error {
	return db.Update(func(tx *bolt.Tx) error {
		users := Users()
		ext_recipients := make([]string, 0)
		_, sender_local := users[e.Sender]

		env_log := log.WithFields(log.Fields{
			"session":      e.Session,
			"sender":       e.Sender,
			"sender_local": sender_local,
			"recipients":   e.Recipients,
			"body_size":    len(e.Body),
			"env_id":       e.ID,
		})

		message_id := ""
		return_path := ""
		msg, err := mail.ReadMessage(bytes.NewReader([]byte(e.Body)))
		if err != nil {
			env_log.Error(err)
		}
		//err != nil means that headers are malformed
		if err != nil || msg.Header.Get("Return-Path") == "" {
			return_path = fmt.Sprintf("Return-Path: <%v>\r\n", e.Sender)
		}
		if err != nil || msg.Header.Get("Message-ID") == "" {
			return_path = fmt.Sprintf("Message-ID: <%v@%v>\r\n", e.ID, config.GetString("server.name"))
		}
		//TODO: check DSN (Disposition-Notification-To)
		headers := message_id + return_path
		if err != nil {
			//MIME error: missing headers
			headers += "\r\n"
		}

		for _, rcpt := range e.Recipients {
			_, rcpt_local := users[rcpt]
			r_env := &Envelope{}
			*r_env = *e

			//if recipient is local, add envelope to their inbox
			if rcpt_local || config.GetBool("db.save_all_mail") {
				delivered_to := fmt.Sprintf("Delivered-To: %v\r\n", rcpt)
				r_headers := delivered_to + headers
				r_env.Body = r_headers + r_env.Body

				b, err := UserBucket(tx, rcpt)
				inbound := b.Bucket([]byte("inbound"))

				env_json, err := json.Marshal(r_env)
				if err != nil {
					log.Fatal(err)
				}

				err = inbound.Put([]byte(r_env.ID), []byte(env_json))
				if err != nil {
					env_log.Error(err)
					return fmt.Errorf("failed to queue")
				}
				env_log.Info("Saved envelope to DB")
			}

			if !rcpt_local {
				ext_recipients = append(ext_recipients, rcpt)
			}
		}

		env_log = env_log.WithField("ext_recipients", ext_recipients)

		//if sender is local and there are external recipients, add envelope to sender outbox
		if sender_local && len(ext_recipients) > 0 {
			if !e.OutboundAllowed {
				env_log.Error("Blocked outbound mail")
				return fmt.Errorf("outbound mail not allowed")
			}

			b, err := UserBucket(tx, e.Sender)
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
			err = outbound.Put([]byte(e.ID), []byte(env_json))
			if err != nil {
				env_log.Error("Couldn't access sender box: ", err)
				return fmt.Errorf("couldn't access sender box")
			}
		}

		return nil
	})
}

func Stat(username string) (int, int) {
	size := 0
	cnt := 0
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(username))
		inbound := b.Bucket([]byte("inbound"))
		err := inbound.ForEach(func(k, v []byte) error {
			env := &Envelope{}
			err := json.Unmarshal(v, &env)
			if err != nil {
				log.Error(err)
				return nil
			}
			cnt++
			size += len(env.Body)
			return nil
		})
		if err != nil {
			log.Error(err)
			return nil
		}
		return nil
	})
	return cnt, size
}

func List(username string) []*Envelope {
	envs := make([]*Envelope, 0)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(username))
		inbound := b.Bucket([]byte("inbound"))
		inbound.ForEach(func(k, v []byte) error {
			env := &Envelope{}
			err := json.Unmarshal(v, &env)
			if err != nil {
				log.Error(err)
				return nil
			}
			envs = append(envs, env)
			return nil
		})
		return nil
	})
	return envs
}

func Sweep() []*Envelope {
	log.Info("Starting outbound sweep")
	envs := make([]*Envelope, 0)
	db.Update(func(tx *bolt.Tx) error {
		for user, _ := range Users() {
			b := tx.Bucket([]byte(user))
			ob := b.Bucket([]byte("outbound"))
			ob.ForEach(func(k, v []byte) error {
				env := &Envelope{}
				err := json.Unmarshal(v, env)
				if err != nil {
					log.WithFields(log.Fields{
						"user":   user,
						"bucket": "outbound",
						"id":     string(k),
					}).Error("Couldn't unmarshal envelope")
				} else {
					//status must be Pending and NextDeliverTime date must have passed
					if env.Status != Pending || !env.NextDeliverTime.Before(time.Now()) {
						return nil
					}
					envs = append(envs, env)
				}
				return nil
			})
		}
		return nil
	})
	return envs
}
