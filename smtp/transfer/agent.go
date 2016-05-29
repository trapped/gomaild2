package transfer

import (
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	. "github.com/trapped/gomaild2/db"
	. "github.com/trapped/gomaild2/structs"
	"time"
)

type Agent struct {
}

var pipeline chan *Envelope = make(chan *Envelope)

func worker() {
	for {
		env := <-pipeline
		log.Info("sending to ", env.Recipients)
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
