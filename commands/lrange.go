package commands

import (
    "fmt"
    "net"

    srv "github.com/codecrafters-io/redis-starter-go/server"
)

func LRange(key string, start, stop int, con net.Conn) {
    v, ok := srv.MEM()[key]
    val, ok2 := v.Data.([]string)
    if !ok || !ok2 || len(val) == 0 {
        fmt.Fprint(con, "*0\r\n")
        return
    }
    if start < 0 {
        start = len(val) + start
        if start < 0 {
            start = 0
        }
    }
    if stop < 0 {
        stop = len(val) + stop
        if stop < 0 {
            stop = 0
        }
    }
    stop++
    if start > len(val) {
        fmt.Fprint(con, "*0\r\n")
        return
    }
    if stop > len(val) {
        stop = len(val)
    }
    if start > stop {
        fmt.Fprint(con, "*0\r\n")
        return
    }
    fmt.Fprintf(con, "*%d\r\n", stop-start)
    for i := start; i < stop; i++ {
        fmt.Fprintf(con, "$%d\r\n%s\r\n", len(val[i]), val[i])
    }
}
