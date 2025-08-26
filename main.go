package main

import (
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/commands"
	srv "github.com/codecrafters-io/redis-starter-go/server"
)

// + inf
const SeqInfinity int64 = 1<<63 - 1

func main() {
	fmt.Println("Starting Redis server...")
	srv.InitServer()

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	fmt.Println("Server listening on port 6379")
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(con net.Conn) {
	defer con.Close()

	b := make([]byte, 9999)

	for {
		numBytes, err := con.Read(b)
		if err != nil {
			fmt.Println("Error reading:", err)
			return
		}

		parts := strings.Split(string(b[:numBytes]), "\r\n")
		args := extractCMD(parts)
		if len(args) == 0 {
			con.Write([]byte("-ERR empty command\r\n"))
			continue
		}
		fmt.Println("User Command: ", args, " Len:", len(args))

		switch strings.ToUpper(args[0]) {
		case "PING":
			commands.Ping(con)
		case "ECHO":
			if len(args) < 2 {
				con.Write([]byte("-ERR wrong number of arguments for 'ECHO'\r\n"))
				continue
			}
			commands.Echo(args[1], con)
		case "QUIT":
			return
		case "SET":
			if len(args) < 3 {
				con.Write([]byte("-ERR wrong number of arguments for 'SET'\r\n"))
				continue
			}
			key, value := args[1], args[2]
			commands.Set(key, value, con, args[3:]...)
		case "GET":
			if len(args) < 2 {
				con.Write([]byte("-ERR wrong number of arguments for 'GET'\r\n"))
				continue
			}
			commands.Get(args[1], con)
		case "RPUSH":
			if len(args) < 3 {
				con.Write([]byte("-ERR wrong number of arguments for 'RPUSH'\r\n"))
				continue
			}
			key := args[1]
			vals := args[2:]
			commands.RPush(key, toList(vals), con)
		case "LPUSH":
			if len(args) < 3 {
				con.Write([]byte("-ERR wrong number of arguments for 'LPUSH'\r\n"))
				continue
			}
			key := args[1]
			vals := args[2:]
			commands.LPush(key, toList(vals), con)
		case "LLEN":
			if len(args) < 2 {
				con.Write([]byte("-ERR wrong number of arguments for 'LLEN'\r\n"))
				continue
			}
			commands.LLen(args[1], con)
		case "LPOP":
			if len(args) < 2 {
				con.Write([]byte("-ERR wrong number of arguments for 'LPOP'\r\n"))
				continue
			}
			key := args[1]
			count := 0
			if len(args) >= 3 {
				if n, err := strconv.Atoi(args[2]); err == nil {
					count = n
				}
			}
			commands.LPop(key, count, con)
		case "LRANGE":
			if len(args) < 4 {
				con.Write([]byte("-ERR wrong number of arguments for 'LRANGE'\r\n"))
				continue
			}
			key := args[1]
			start, _ := strconv.Atoi(args[2])
			end, _ := strconv.Atoi(args[3])
			commands.LRange(key, start, end, con)
		case "BLPOP":
			if len(args) < 3 {
				con.Write([]byte("-ERR wrong number of arguments for 'BLPOP'\r\n"))
				continue
			}
			key := args[1]
			dur, _ := time.ParseDuration(args[2] + "s")
			commands.BLPop(key, dur, con)
		case "TYPE":
			if len(args) < 2 {
				con.Write([]byte("-ERR wrong number of arguments for 'TYPE'\r\n"))
				continue
			}
			commands.Type(args[1], con)
		case "XADD":
			if len(args) < 2 {
				con.Write([]byte("-ERR wrong number of arguments for 'XADD'\r\n"))
				continue
			}
			key := args[1]
			id := args[2]
			data := toStream(args[3:])
			commands.Xadd(key, id, data, con)
		case "XRANGE":
			// User Command:  [XRANGE some_key 1526985054069 1526985054079]  Len: 4

			if len(args) < 3 {
				con.Write([]byte("-ERR wrong number of arguments for 'XRANGE'\r\n"))
				continue
			}
			key := args[1]
			start, startSeq := toTimeSeq(args[2])
			end, endSeq := toTimeSeq(args[3])

			if args[3] == "+" {
				end, endSeq = SeqInfinity, SeqInfinity

			}
			if args[2] == "-" {
				start, startSeq = 0, 0
			}

			commands.Xrange(key, start, startSeq, end, endSeq, con)
		case "XREAD":
			// XREAD [BLOCK milliseconds] STREAMS key [key ...] id [id ...]
			// User Command:  [XREAD streams some_key 1526985054069-0]  Len: 4

			idx := slices.Index(args, "streams")

			start := idx + 1
			rem := len(args) - start

			mid := start + rem/2
			keys := args[start:mid]
			ids := args[mid:]

			pairs := make([][]any, 0, len(keys))
			for i := range keys {
				time, seq := toTimeSeq(ids[i])
				pairs = append(pairs, []any{keys[i], time, seq})
			}
			commands.Xread(pairs, con)
		default:
			con.Write([]byte("-ERR unknown command\r\n"))
		}
	}
}

func extractCMD(parts []string) []string {
	vals := make([]string, 0)
	for i := 0; i+1 < len(parts); i++ {
		if strings.HasPrefix(parts[i], "$") {
			v := parts[i+1]
			if v != "" {
				vals = append(vals, v)
			}
		}
	}
	return vals
}

func toList(vals []string) []string {
	out := make([]string, 0, len(vals)*2)
	for _, v := range vals {
		out = append(out, "$"+strconv.Itoa(len(v)), v)
	}
	fmt.Println(out)
	return out
}

func toStream(vals []string) []srv.Field {
	out := []srv.Field{}
	for i := 0; i < len(vals); i += 2 {
		out = append(out, srv.Field{Key: vals[i], Value: vals[i+1]})
	}

	return out
}

func toTimeSeq(s string) (int64, int64) {
	var seq int64
	seq = 0
	parts := strings.Split(s, "-")
	val, _ := strconv.ParseInt(parts[0], 10, 64)
	if len(parts) == 2 && parts[1] != "" {
		seq, _ = strconv.ParseInt(parts[1], 10, 64)
	}
	return val, seq
}
