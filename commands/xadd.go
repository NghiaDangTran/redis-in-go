package commands

import (
	"fmt"
	"net"

	srv "github.com/codecrafters-io/redis-starter-go/server"
)

func Xadd(k string, id string, data map[string]string, con net.Conn) {

	fmt.Fprintf(con, "$%d\r\n%s\r\n", len(id), id)
	srv.MEM()[k] = srv.Value{DataType: srv.Stream, Data: data}

}
