package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/trapped/gomaild2/smtp"
	"io"
	"os"
)

var logpath = "./gomaild2.log"

func init() {
	logfile, err := os.OpenFile(logpath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		log.Fatalf("Couldn't open log file: %v\n", err.Error())
	}
	log.SetOutput(io.MultiWriter(os.Stderr, logfile))
	log.SetFormatter(&log.TextFormatter{ForceColors: false})
}

func main() {
	smtp25 := &smtp.Server{Addr: "0.0.0.0", Port: "25"}
	smtp25.Start()
}
