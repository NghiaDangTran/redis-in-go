package commands

import (
    "fmt"
    "net"
    "strconv"
    "strings"
    "time"

    srv "github.com/codecrafters-io/redis-starter-go/server"
)

// Set sets a string value, optionally with PX expiry.
func Set(k, v string, con net.Conn, argv ...string) {
    srv.MEM()[k] = srv.Value{DataType: srv.String, Data: v}
    fmt.Fprint(con, "+OK\r\n")

    // parse optional PX from additional tokens
    for i := 0; i+1 < len(argv); i++ {
        if strings.EqualFold(argv[i], "PX") {
            if ms, err := strconv.Atoi(argv[i+1]); err == nil {
                go func(key string, d time.Duration) {
                    <-time.After(d)
                    delete(srv.MEM(), key)
                }(k, time.Duration(ms)*time.Millisecond)
            }
        }
    }
}
