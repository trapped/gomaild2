package structs

import (
	"net"
)

type Server interface {
	Start()
	Stop()
}
