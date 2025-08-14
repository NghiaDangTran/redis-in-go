package commands

import (
    "fmt"
    "net"

    srv "github.com/codecrafters-io/redis-starter-go/server"
)

func LPop(k string, count int, con net.Conn) {
    v, ok := srv.MEM()[k]
    val, ok2 := v.Data.([]string)
    if !ok || !ok2 || len(val) == 0 {
        fmt.Fprint(con, "$-1\r\n")
        return
    }
    if count <= 0 {
        s := val[0]
        srv.MEM()[k] = srv.Value{DataType: srv.List, Data: val[1:]}
        fmt.Fprintf(con, "$%d\r\n%s\r\n", len(s), s)
        return
    }
    if count > len(val) {
        count = len(val)
    }
    fmt.Fprintf(con, "*%d\r\n", count)
    for i := 0; i < count; i++ {
        s := val[i]
        fmt.Fprintf(con, "$%d\r\n%s\r\n", len(s), s)
    }
    srv.MEM()[k] = srv.Value{DataType: srv.List, Data: val[count:]}
}
