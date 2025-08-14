package commands

import (
	"fmt"
	"net"

	srv "github.com/codecrafters-io/redis-starter-go/server"
)

func Type(k string, con net.Conn) {
	v, ok := srv.MEM()[k]
	if !ok {
		fmt.Fprint(con, "+none\r\n")
		return
	}
	switch v.DataType {
	case srv.String:
		fmt.Fprint(con, "+string\r\n")
	case srv.List:
		fmt.Fprint(con, "+list\r\n")
	case srv.Stream:
		fmt.Fprint(con, "+stream\r\n")
	default:
		fmt.Fprint(con, "+none\r\n")
	}
}
