package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/commands"
	srv "github.com/codecrafters-io/redis-starter-go/server"
)

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

func toStream(vals []string) map[string]string {
	out := make(map[string]string)
	for i := 0; i < len(vals); i += 2 {
		out[vals[i]] = vals[i+1]
	}
	return out
}
