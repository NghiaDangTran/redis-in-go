package commands

import (
    "fmt"
    "net"
)

func Echo(msg string, con net.Conn) {
    fmt.Fprintf(con, "+%s\r\n", msg)
}

