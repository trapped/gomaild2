package main

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	"github.com/trapped/gomaild2/db"
	"github.com/trapped/gomaild2/pop3"
	"github.com/trapped/gomaild2/smtp"
	"io"
	"os"
)

func initlog() {
	logfile, err := os.OpenFile(config.GetString("log.path"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		log.Fatalf("Couldn't open log file: %v", err.Error())
	}
	log.SetOutput(io.MultiWriter(os.Stderr, logfile))
	log.SetFormatter(&log.TextFormatter{ForceColors: false})
	log.SetLevel(log.DebugLevel)
}

func initconfig() {
	//logging
	config.SetDefault("log.path", "gomaild2.log")
	//server
	config.SetDefault("server.name", "localhost")
	//smtp
	config.SetDefault("server.smtp.mta.address", "0.0.0.0")
	config.SetDefault("server.smtp.msa.address", "0.0.0.0")
	config.SetDefault("server.smtp.mta.ports", []int{25})
	config.SetDefault("server.smtp.msa.ports", []int{587})
	config.SetDefault("server.smtp.mta.require_auth", false)
	config.SetDefault("server.smtp.msa.require_auth", true)
	config.SetDefault("server.smtp.msa.outbound", false)
	config.SetDefault("server.smtp.msa.outbound", true)
	//password encryption
	config.BindEnv("pw_encryption") //AES256 GCM key to decrypt passwords from config file
	config.SetDefault("pw_encryption", "")
	//db
	config.SetDefault("db.path", "gomaild2.db")
	config.SetDefault("db.accept_all_mail", true)
	//tls
	config.SetDefault("tls.enabled", false)
	config.SetDefault("tls.certificate", "")
	config.SetDefault("tls.key", "")
	//meta
	config.SetDefault("config.loaded", false)
	config.SetDefault("encryption.loaded", false)
	//read config
	config.SetConfigName("config")
	config.AddConfigPath(".")
	err := config.ReadInConfig()
	if err != nil {
		log.Fatalf("Config error: %v", err.Error())
	}
	config.WatchConfig()
	config.OnConfigChange(func(e fsnotify.Event) {
		log.Info("Config file changed: " + e.Name)
		db.Reopen()
	})
	config.Set("config.loaded", true)
	log.Debug("Config ready")
}

func init() {
	initconfig()
	initlog()
}

func main() {
	run := make(<-chan struct{})
	db.Open()
	defer db.Close()
	//smtp
	server := config.Sub("server")
	for name, _ := range server.GetStringMap("smtp") {
		srv := server.Sub("smtp").Sub(name)
		for _, port := range srv.GetStringSlice("ports") {
			smtp := &smtp.Server{
				Addr:        srv.GetString("address"),
				Port:        port,
				RequireAuth: srv.GetBool("require_auth"),
				Outbound:    srv.GetBool("outbound"),
			}
			go smtp.Start()
		}
	}
	//pop3
	for name, _ := range server.GetStringMap("pop3") {
		srv := server.Sub("pop3").Sub(name)
		for _, port := range srv.GetStringSlice("ports") {
			pop3 := &pop3.Server{
				Addr: srv.GetString("address"),
				Port: port,
			}
			go pop3.Start()
		}
	}
	<-run
}
