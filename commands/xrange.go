package commands

import (
	"fmt"
	"net"
	"strconv"

	srv "github.com/codecrafters-io/redis-starter-go/server"
)

func Xrange(k string, startTime, startSeq, endTime, endSeq int64, con net.Conn) {
	mem := srv.MEM()
	item, ok := mem[k]
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
		if e.Time > startTime || (e.Time == startTime && int64(e.Sequence) >= startSeq) {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		fmt.Fprintf(con, "*0\r\n")
		return
	}

	endIdx := -1
	for i := len(sl) - 1; i >= startIdx; i-- {
		e := sl[i]
		if e.Time < endTime || (e.Time == endTime && int64(e.Sequence) <= endSeq) {
			endIdx = i
			break
		}
	}
	if endIdx == -1 || endIdx < startIdx {
		fmt.Fprintf(con, "*0\r\n")
		return
	}

	res := sl[startIdx : endIdx+1]

	fmt.Fprintf(con, "*%d\r\n", len(res))

	for _, entry := range res {
		fmt.Fprintf(con, "*2\r\n")

		id := strconv.FormatInt(entry.Time, 10) + "-" + strconv.Itoa(entry.Sequence)
		fmt.Fprintf(con, "$%d\r\n%s\r\n", len(id), id)

		fmt.Fprintf(con, "*%d\r\n", 2*len(entry.Fields))
		for _, f := range entry.Fields {
			fmt.Fprintf(con, "$%d\r\n%s\r\n", len(f.Key), f.Key)
			fmt.Fprintf(con, "$%d\r\n%s\r\n", len(f.Value), f.Value)
		}
	}
}
