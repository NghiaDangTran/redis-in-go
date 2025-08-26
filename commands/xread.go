package commands

import (
	"fmt"
	"net"
	"strconv"

	srv "github.com/codecrafters-io/redis-starter-go/server"
)

func Xread(pairs [][]any, con net.Conn) {
	mem := srv.MEM()

	fmt.Println(pairs)
	fmt.Fprintf(con, "*%d\r\n", len(pairs))

	for i := range pairs {
		key, time, seq := pairs[i][0].(string), pairs[i][1].(int64), pairs[i][2].(int64)

		item, ok := mem[key]
		if !ok {
			fmt.Fprintf(con, "*0\r\n")
			return
		}

		var sd srv.StreamData
		switch v := item.Data.(type) {
		case srv.StreamData:
			sd = v
		case *srv.StreamData:
			sd = *v
		default:
			fmt.Fprintf(con, "*0\r\n")
			return
		}

		sl := sd.StreamList
		if len(sl) == 0 {
			fmt.Fprintf(con, "*0\r\n")
			return
		}

		startIdx := -1
		for i, e := range sl {
			if e.Time > time || (e.Time == time && int64(e.Sequence) >= seq) {
				startIdx = i
				break
			}
		}
		if startIdx == -1 {
			fmt.Fprintf(con, "*0\r\n")
			return
		}
		fmt.Fprintf(con, "*2\r\n")
		fmt.Fprintf(con, "$%d\r\n%s\r\n", len(key), key)
		fmt.Fprintf(con, "*1\r\n")
		fmt.Fprintf(con, "*2\r\n")
		res := sl[startIdx]

		id := strconv.FormatInt(res.Time, 10) + "-" + strconv.Itoa(res.Sequence)
		fmt.Fprintf(con, "$%d\r\n%s\r\n", len(id), id)
		fmt.Fprintf(con, "*%d\r\n", 2*len(res.Fields))
		for _, f := range res.Fields {
			fmt.Fprintf(con, "$%d\r\n%s\r\n", len(f.Key), f.Key)
			fmt.Fprintf(con, "$%d\r\n%s\r\n", len(f.Value), f.Value)
		}
	}

}
