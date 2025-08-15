package commands

import (
	"fmt"
	"net"
)

func Ping(con net.Conn) {
	fmt.Fprint(con, "+PONG\r\n")
}
