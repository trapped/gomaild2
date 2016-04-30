package db

import (
	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	"strings"
)

var (
	db *bolt.DB
)

func init() {
	go initEncryption()
}

func Users() map[string]string {
	users := make(map[string]string)
	for domain, _ := range config.GetStringMap("domains") {
		for _, userstring := range config.Sub("domains").Sub(domain).GetStringSlice("users") {
			us_split := strings.Split(userstring, "@")
			username, password := us_split[0], decryptPassword(us_split[1])
			users[username+"@"+domain] = password
		}
	}
	log.Debug(users)
	return users
}

func userBucket(tx *bolt.Tx, user string) (*bolt.Bucket, error) {
	b, err := tx.CreateBucketIfNotExists([]byte(user))
	_, err = b.CreateBucketIfNotExists([]byte("inbound"))
	_, err = b.CreateBucketIfNotExists([]byte("outbound"))
	return b, err
}

func Open() {
	d, err := bolt.Open(config.GetString("db.path"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	d.Update(func(tx *bolt.Tx) error {
		for user, _ := range Users() {
			_, err := userBucket(tx, user)
			if err != nil {
				return err
			}
			log.WithFields(log.Fields{
				"user": user,
			}).Info("Mailbox loaded")
		}
		return nil
	})
	db = d
}

func Close() {
	if db != nil {
		db.Close()
		db = nil
	} else {
		log.Warn("db.Close called while db was already closed")
	}
}

func Reopen() {
	go Close()
	Open()
}
