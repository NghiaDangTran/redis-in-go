package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	srv "github.com/codecrafters-io/redis-starter-go/server"
)

func Xadd(k string, id string, data map[string]string, con net.Conn) {

	mem := srv.MEM()
	v, ok := mem[k]

	newTime, newSeq, err := extractID(id, v, ok, con)
	if err != nil {
		return
	}

	if !ok {
		mem[k] = srv.Value{
			DataType: srv.Stream,
			Data: &srv.StreamData{
				TimeMap: map[int64]int{newTime: newSeq},
				StreamList: []srv.StreamEntry{
					{Time: newTime, Sequence: newSeq, Fields: data},
				}},
		}

	} else {
		sd := v.Data.(*srv.StreamData)
		sd.StreamList = append(sd.StreamList, srv.StreamEntry{Time: newTime, Sequence: newSeq, Fields: data})

		mem[k] = v
	}
	strLen := fmt.Sprintf("%d-%d", newTime, newSeq)
	fmt.Fprintf(con, "$%d\r\n%s\r\n", len(strLen), strLen)

}
func extractID(id string, v srv.Value, ok bool, con net.Conn) (int64, int, error) {
	if strings.Contains(id, "*") {
		parts := strings.SplitN(id, "-", 2)
		if parts[0] == "*" {
			if ok {
				// generate current time
				newTime := time.Now().Unix()
				d := v.Data.(*srv.StreamData)

				if _, exist := d.TimeMap[newTime]; !exist {
					d.TimeMap[newTime] = 0
				} else {
					d.TimeMap[newTime]++
				}

				return newTime, d.TimeMap[newTime], nil
			} else {
				// generate time and start at zero
				newTime := time.Now().Unix()
				return newTime, 0, nil
			}

		}
		if parts[1] == "*" {
			newTime, _ := strconv.ParseInt(parts[0], 10, 64)

			if ok {
				d := v.Data.(*srv.StreamData)
				if _, exist := d.TimeMap[newTime]; !exist {
					d.TimeMap[newTime] = 0
				} else {
					d.TimeMap[newTime]++
				}

				return newTime, d.TimeMap[newTime], nil

			} else {
				if newTime == 0 {
					return 0, 1, nil
				}

				return newTime, 0, nil

			}

		}
	}

	parts := strings.SplitN(id, "-", 2)

	newTime, _ := strconv.ParseInt(parts[0], 10, 64)
	newSeq, _ := strconv.Atoi(parts[1])
	if newTime <= 0 && newSeq <= 0 {
		fmt.Fprintf(con, "-ERR The ID specified in XADD must be greater than 0-0\r\n")
		return 0, 0, fmt.Errorf("xadd id not increasing")
	}
	if ok {
		d := v.Data.(*srv.StreamData)
		if n := len(d.StreamList); n > 0 {
			last := d.StreamList[n-1]
			if newTime < last.Time || (newTime == last.Time && newSeq <= last.Sequence) {
				fmt.Fprintf(con, "-ERR The ID specified in XADD is equal or smaller than the target stream top item\r\n")
				return 0, 0, fmt.Errorf("xadd id not increasing")
			}
		}
		d.TimeMap[newTime] = newSeq
	}
	return newTime, newSeq, nil
}
