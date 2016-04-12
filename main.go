package main

import (
	"trapped/gomaild2/smtp"
)

func main() {
	smtp25 := &smtp.Server{Addr: "0.0.0.0", Port: "25"}
	smtp25.Start()
}
