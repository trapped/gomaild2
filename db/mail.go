package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	. "github.com/trapped/gomaild2/structs"
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

func (e *Envelope) Save() (string, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := e.save(tx)
		return err
	})
	return e.ID, err
}

func (e *Envelope) save(tx *bolt.Tx) (string, error) {
	if e.ID != "" {
		//envelope has already been previously saved, update only
		log.Info("Updating envelope")
		users := Users()
		if e.Type == Inbound {
			//if the type is inbound it has been fetched from an inbound box (likely from pop3)
			for _, rcpt := range e.Recipients {
				_, rcpt_local := users[rcpt]
				if rcpt_local || config.GetBool("db.save_all_mail") {
					b, err := UserBucket(tx, rcpt)
					if err != nil {
						log.WithField("err", err).Error("Couldn't open user box")
						return e.ID, fmt.Errorf("couldn't access user box")
					}
					inbound := b.Bucket([]byte("inbound"))

					env_json, err := json.Marshal(e)
					if err != nil {
						log.WithField("err", err).Error("Couldn't marshal envelope")
						return e.ID, fmt.Errorf("failed to marshal")
					}

					err = inbound.Put([]byte(e.ID), []byte(env_json))
					if err != nil {
						log.WithField("err", err).Error("Couldn't save envelope")
						return e.ID, fmt.Errorf("failed to update")
					}
				}
			}
		} else if e.Type == Outbound {
			//if the type is outbound it has been fetched from an inbound box (likely from transfer)
			_, sender_local := users[e.Sender]
			if sender_local {
				b, err := UserBucket(tx, e.Sender)
				if err != nil {
					log.WithField("err", err).Error("Couldn't open user box")
					return e.ID, fmt.Errorf("couldn't access user box")
				}

				env_json, err := json.Marshal(e)
				if err != nil {
					log.WithField("err", err).Error("Couldn't marshal envelope")
					return e.ID, fmt.Errorf("failed to marshal")
				}

				outbound := b.Bucket([]byte("outbound"))
				err = outbound.Put([]byte(e.ID), []byte(env_json))
				if err != nil {
					log.WithField("err", err).Error("Couldn't save envelope")
					return e.ID, fmt.Errorf("failed to update")
				}
			}
		}
		return e.ID, nil
	}
	//new envelope, assign ID, add headers, classify (inbound/outbound, etc.) and save
	e.ID = e.Session + "." + SessionID(12)

	users := Users()
	ext_recipients := make([]string, 0)
	_, sender_local := users[e.Sender]

	env_log := log.WithFields(log.Fields{
		"session":      e.Session,
		"sender":       e.Sender,
		"sender_local": sender_local,
		"recipients":   e.Recipients,
		"body_size":    len(e.Body),
		"env":          e.ID,
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
			if err != nil {
				env_log.WithField("err", err).Error("Couldn't open user box")
				return e.ID, fmt.Errorf("couldn't access user box")
			}
			inbound := b.Bucket([]byte("inbound"))

			env_json, err := json.Marshal(r_env)
			if err != nil {
				env_log.WithField("err", err).Error("Couldn't marshal envelope")
				return e.ID, fmt.Errorf("failed to marshal")
			}

			err = inbound.Put([]byte(r_env.ID), []byte(env_json))
			if err != nil {
				env_log.WithField("err", err).Error("Couldn't save envelope")
				return e.ID, fmt.Errorf("failed to save")
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
			return e.ID, fmt.Errorf("outbound mail not allowed")
		}

		b, err := UserBucket(tx, e.Sender)
		if err != nil {
			env_log.WithField("err", err).Error("Couldn't open user box")
			return e.ID, fmt.Errorf("couldn't access user box")
		}

		//convert envelope to Outbound
		o_env := &Envelope{}
		*o_env = *e
		o_env.Type = Outbound
		o_env.Status = Pending
		o_env.Recipients = ext_recipients

		env_json, err := json.Marshal(o_env)
		if err != nil {
			env_log.WithField("err", err).Error("Couldn't marshal envelope")
			return e.ID, fmt.Errorf("failed to marshal")
		}

		outbound := b.Bucket([]byte("outbound"))
		err = outbound.Put([]byte(e.ID), []byte(env_json))
		if err != nil {
			env_log.WithField("err", err).Error("Couldn't save envelope")
			return e.ID, fmt.Errorf("failed to save")
		}
	}

	return e.ID, err
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
						"err":    err,
						"user":   user,
						"bucket": "outbound",
						"env":    string(k),
					}).Error("Couldn't unmarshal envelope")
				} else {
					//status must be Pending and NextDeliverTime date must have passed
					if env.Status != Pending || !env.NextDeliverTime.Before(time.Now()) {
						return nil
					}
					env.Status = Assigned
					_, err := env.save(tx)
					if err != nil {
						log.WithFields(log.Fields{
							"err":    err,
							"user":   user,
							"bucket": "outbound",
							"env":    string(k),
						}).Error("Couldn't assign envelope")
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
