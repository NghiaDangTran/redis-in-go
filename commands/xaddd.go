package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	srv "github.com/codecrafters-io/redis-starter-go/server"
)

func Xadd(k string, id string, data map[string]string, con net.Conn) {

	mem := srv.MEM()

	newTime, newSeq := extractID(id)
	if newTime <= 0 && newSeq <= 0 {
		fmt.Fprintf(con, "-ERR The ID specified in XADD must be greater than 0-0\r\n")
		return
	}
	if v, ok := mem[k]; !ok {

		mem[k] = srv.Value{
			DataType: srv.Stream,
			Data: []srv.StreamEntry{
				{ID: id, Fields: data},
			},
		}

	} else {
		lastVal := v.Data.([]srv.StreamEntry)
		lastTime, lastSeq := extractID(lastVal[len(lastVal)-1].ID)
		fmt.Println(lastTime, newTime, lastSeq, newSeq)
		if newTime < lastTime || (newTime == lastTime && newSeq <= lastSeq) {
			fmt.Fprintf(con, "-ERR The ID specified in XADD is equal or smaller than the target stream top item\r\n")
			return
		}

		s := append(v.Data.([]srv.StreamEntry), srv.StreamEntry{ID: id, Fields: data})
		v.Data = s
		mem[k] = v
	}

	fmt.Fprintf(con, "$%d\r\n%s\r\n", len(id), id)

}
func extractID(id string) (int, int) {
	parts := strings.SplitN(id, "-", 2)
	val1, _ := strconv.Atoi(parts[0])
	val2, _ := strconv.Atoi(parts[1])
	return val1, val2
}
