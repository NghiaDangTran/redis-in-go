package commands

import (
	"fmt"
	"net"

	srv "github.com/codecrafters-io/redis-starter-go/server"
)

func LPush(k string, values []string, con net.Conn) {
	toAdd := make([]string, 0, len(values)/2)
	for i := len(values) - 1; i > 0; i -= 2 {
		toAdd = append(toAdd, values[i])
	}

	var cur []string
	if v, ok := srv.MEM()[k]; ok {
		cur, _ = v.Data.([]string)
	}
	cur = append(toAdd, cur...)
	srv.MEM()[k] = srv.Value{DataType: srv.List, Data: cur}
	fmt.Fprintf(con, ":%d\r\n", len(cur))
}
