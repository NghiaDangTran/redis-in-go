package commands

import (
	"fmt"
	"net"

	srv "github.com/codecrafters-io/redis-starter-go/server"
)

func LLen(k string, con net.Conn) {
	if v, ok := srv.MEM()[k]; ok {
		if val, ok2 := v.Data.([]string); ok2 {
			fmt.Fprintf(con, ":%d\r\n", len(val))
			return
		}
	}
	fmt.Fprint(con, ":0\r\n")
}
