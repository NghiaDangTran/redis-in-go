package commands

import (
    "fmt"
    "net"

    srv "github.com/codecrafters-io/redis-starter-go/server"
)

func Get(k string, con net.Conn) {
    if v, ok := srv.MEM()[k]; ok {
        if v.DataType == srv.String {
            if s, ok2 := v.Data.(string); ok2 {
                fmt.Fprintf(con, "$%d\r\n%s\r\n", len(s), s)
                return
            }
        }
    }
    fmt.Fprint(con, "$-1\r\n")
}
