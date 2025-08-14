package commands

import (
    "fmt"
    "net"
    "time"

    srv "github.com/codecrafters-io/redis-starter-go/server"
    "github.com/codecrafters-io/redis-starter-go/utils"
)

func BLPop(k string, wait time.Duration, con net.Conn) {
    id := utils.TimeHash()
    srv.CHANS()[id] = make(chan bool)
    defer delete(srv.CHANS(), id)

    timeout := make(chan bool, 1)
    if wait > 0 {
        go func() {
            <-time.After(wait)
            timeout <- true
        }()
    }
    for {
        select {
        case <-srv.CHANS()[id]:
            if v, ok := srv.MEM()[k]; ok {
                if val, ok2 := v.Data.([]string); ok2 && len(val) > 0 {
                    s := val[0]
                    srv.MEM()[k] = srv.Value{DataType: srv.List, Data: val[1:]}
                    fmt.Fprintf(con, "*2\r\n")
                    fmt.Fprintf(con, "$%d\r\n%s\r\n", len(k), k)
                    fmt.Fprintf(con, "$%d\r\n%s\r\n", len(s), s)
                    return
                }
            }
        case <-timeout:
            fmt.Fprint(con, "$-1\r\n")
            return
        }
    }
}
