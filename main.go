package main

import (
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	"github.com/trapped/gomaild2/smtp"
	"io"
	"os"
	"time"
)

func initlog() {
	logfile, err := os.OpenFile(config.GetString("log.path"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		log.Fatalf("Couldn't open log file: %v", err.Error())
	}
	log.SetOutput(io.MultiWriter(os.Stderr, logfile))
	log.SetFormatter(&log.TextFormatter{ForceColors: false})
}

func initconfig() {
	config.SetConfigName("config")
	config.AddConfigPath(".")
	err := config.ReadInConfig()
	if err != nil {
		log.Fatalf("Config error: %v", err.Error())
	}
	config.WatchConfig()
	//logging
	config.SetDefault("log.path", "./gomaild2.log")
	//global
	config.SetDefault("server.name", "localhost")
	//smtp
	config.SetDefault("server.smtp.address", "0.0.0.0")
	config.SetDefault("server.smtp.ports", []int{25})
}

func init() {
	initconfig()
	initlog()
}

func main() {
	for _, port := range config.GetStringSlice("server.smtp.ports") {
		smtp := &smtp.Server{
			Addr: config.GetString("server.smtp.address"),
			Port: port,
		}
		go smtp.Start()
	}
	for {
		time.Sleep(10 * time.Second)
	}
}
