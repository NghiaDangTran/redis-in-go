package commands

import (
	"fmt"
	"net"
	"sort"

	srv "github.com/codecrafters-io/redis-starter-go/server"
)

func RPush(k string, values []string, con net.Conn) {
	toAdd := make([]string, 0, len(values)/2)
	for i := 1; i < len(values); i += 2 {
		toAdd = append(toAdd, values[i])
	}

	var cur []string
	if v, ok := srv.MEM()[k]; ok {
		cur, _ = v.Data.([]string)
	}
	cur = append(cur, toAdd...)
	srv.MEM()[k] = srv.Value{DataType: srv.List, Data: cur}
	fmt.Fprintf(con, ":%d\r\n", len(cur))

	if len(srv.CHANS()) > 0 {
		keys := make([]string, 0, len(srv.CHANS()))
		for id := range srv.CHANS() {
			keys = append(keys, id)
		}
		sort.Sort(sort.Reverse(sort.StringSlice(keys)))
		for _, id := range keys {
			srv.CHANS()[id] <- true
		}
	}
}
