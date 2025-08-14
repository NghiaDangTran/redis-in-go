package commands

import (
	"fmt"
	"net"
)

func TYPE(k string, con net.Conn, mem map[string]any) {
	val, ok := mem[k]

	if !ok {
		fmt.Fprintf(con, "+none\r\n")
	}

	switch val.(type) {
	case []string:
		fmt.Fprintf(con, "+list\r\n")

	case string:
		fmt.Fprintf(con, "+string\r\n")

	}

}
