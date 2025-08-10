package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	_ = net.Listen
	_ = os.Exit
)

// avalaibel around 10000 key.
var MEM = make(map[string]any, 10000)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	fmt.Println("Successed to bind to port 6379")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go HandelConnection(conn)
	}
}

func HandelConnection(con net.Conn) {
	defer con.Close()

	b := make([]byte, 9999)

	for {
		numBytes, err := con.Read(b)
		if err != nil {
			fmt.Println("Error reading:", err)

			return
		}

		cmd := strings.Split(string(b[:numBytes]), "\r\n")

		fmt.Printf("User Command: \"%s\", len(%d) \n", cmd, len(cmd))

		switch strings.ToUpper(cmd[2]) {
		case "PING":
			con.Write([]byte("+PONG\r\n"))
		case "ECHO":
			if len(cmd) < 5 {
				con.Write(
					[]byte("-ERR Not enough arguments for ECHO command \r\n"),
				)
			} else {
				message := cmd[4]
				con.Write([]byte(fmt.Sprintf("+%s\r\n", message)))
			}

		case "QUIT":
			return
		case "SET":
			// sample "[*3 $3 set $4 test $2 ok ]"
			if len(cmd) < 6 {
				con.Write(
					[]byte("-ERR Not enough arguments for SET command \r\n"),
				)
			} else {
				key := cmd[4]
				value := cmd[6]
				// with expiry ags  "[*5 $3 SET $3 foo $3 bar $2 px $3 100 ]"
				SET(key, value, con, cmd...)
			}

		case "GET":
			// sample: "[*2 $3 get $2 hi ]"
			if len(cmd) < 4 {
				con.Write(
					[]byte("-ERR Not enough arguments for GET command \r\n"),
				)
			} else {
				key := cmd[4]
				GET(key, con)
			}
		case "RPUSH":
			// sample: "[*3 $5 RPUSH $8 list_key $3 foo ]"
			if len(cmd) < 6 {
				con.Write(
					[]byte("-ERR Not enough arguments for RPUSH command \r\n"),
				)
			} else {
				key := cmd[4]
				value := cmd[5 : len(cmd)-1]
				// with expiry ags  "[*5 $3 SET $3 foo $3 bar $2 px $3 100 ]"\
				// User Command: "[*5 $5 RPUSH $12 another_list $3 foo $3 bar $3 baz ]"
				RPUSH(key, value, con)
			}
		case "LRANGE":
			//  User Command: "[*4 $6 LRANGE $8 list_key $1 0 $1 2 ]"
			// User Command: "[*4 $6 LRANGE $8 list_key $2 -2 $2 -1 ]"
			if len(cmd) < 8 {
				con.Write(
					[]byte("-ERR Not enough arguments for LRANGE command \r\n"),
				)
			} else {
				key := cmd[4]
				start, _ := strconv.Atoi(cmd[6])
				end, _ := strconv.Atoi(cmd[8])
				// with expiry ags  "[*5 $3 SET $3 foo $3 bar $2 px $3 100 ]"\
				// User Command: "[*5 $5 RPUSH $12 another_list $3 foo $3 bar $3 baz ]"
				LRANGE(key, start, end, con)
			}
		case "LPUSH":
			// User Command: "[*5 $5 LPUSH $8 list_key $1 a $1 b $1 c ]"
			if len(cmd) < 6 {
				con.Write(
					[]byte("-ERR Not enough arguments for RPUSH command \r\n"),
				)
			} else {
				key := cmd[4]
				value := cmd[5 : len(cmd)-1]
				// with expiry ags  "[*5 $3 SET $3 foo $3 bar $2 px $3 100 ]"\
				// User Command: "[*5 $5 RPUSH $12 another_list $3 foo $3 bar $3 baz ]"
				LPUSH(key, value, con)
			}
		case "LLEN":
			if len(cmd) < 6 {
				con.Write(
					[]byte("-ERR Not enough arguments for LLEN command \r\n"),
				)
			} else {
				key := cmd[4]
				LLEN(key, con)
			}
		case "LPOP":
			if len(cmd) < 6 {
				con.Write(
					[]byte("-ERR Not enough arguments for LPOP command \r\n"),
				)
			} else {
				key := cmd[4]
				LPOP(key, con)
			}
		default:
			con.Write([]byte("-ERR unknown command\r\n"))
		}
	}
}

func LPOP(k string, con net.Conn) {
	val, ok := MEM[k].([]string)
	if !ok {
		fmt.Fprintf(con, "$%d\r\n", -1)
		return
	}

	fmt.Fprintf(con, "$%d\r\n%s\r\n", len(val[0]), val[0])
	MEM[k] = val[1:]
}
func LLEN(k string, con net.Conn) {
	val, ok := MEM[k].([]string)
	if !ok {
		fmt.Fprintf(con, ":%d\r\n", 0)
		return
	}
	fmt.Fprintf(con, ":%d\r\n", len(val))
}
func LPUSH(k string, v []string, con net.Conn) {

	toAdd := make([]string, len(v)/2)
	// so this make [ "", "" ,""] len of v/2
	// so when you append it
	//  safe make([]string, 0,len(v)/2)
	toAdd = make([]string, 0, len(v)/2)
	for i := len(v) - 1; i > 0; i -= 2 {
		toAdd = append(toAdd, v[i])
	}

	val, _ := MEM[k].([]string)

	MEM[k] = append(toAdd, val...)

	fmt.Fprintf(con, ":%d\r\n", len(MEM[k].([]string)))

}

func LRANGE(key string, start int, end int, con net.Conn) {

	fmt.Println(start, end)
	val, ok := MEM[key].([]string)
	if !ok {
		fmt.Fprintf(con, "*%d\r\n", 0)
		return
	}
	if start < 0 {
		if -start > len(val) {
			start = 0
		} else {
			start = len(val) + start

		}

	}
	if end < 0 {
		if -end > len(val) {
			end = 0
		} else {
			end = len(val) + end

		}
	}

	end = end + 1

	if start > len(val) {
		fmt.Fprintf(con, "*%d\r\n", 0)
		return

	}
	if end > len(val) {
		end = len(val)
	}

	if start > end {
		fmt.Fprintf(con, "*%d\r\n", 0)
		return

	}

	fmt.Fprintf(con, "*%d\r\n", end-start)

	for start < end {
		fmt.Fprintf(con, "$%d\r\n", len(val[start]))
		fmt.Fprintf(con, "%s\r\n", val[start])
		start += 1
	}
}
func RPUSH(k string, v []string, con net.Conn) {

	toAdd := make([]string, len(v)/2)
	// so this make [ "", "" ,""] len of v/2
	// so when you append it
	//  safe make([]string, 0,len(v)/2)
	toAdd = make([]string, 0, len(v)/2)
	for i := 1; i < len(v); i += 2 {
		toAdd = append(toAdd, v[i])
	}

	val, _ := MEM[k].([]string)

	MEM[k] = append(val, toAdd...)

	fmt.Fprintf(con, ":%d\r\n", len(MEM[k].([]string)))

}

func SET(k, v string, con net.Conn, agr ...string) {
	MEM[k] = v

	con.Write([]byte("+OK\r\n"))

	if len(agr) > 8 && strings.ToUpper(agr[8]) == "PX" {
		ms, err := strconv.Atoi(agr[10])
		if err != nil {
			fmt.Println("Invalid PX value:", agr[1])

			return
		}

		go func() {
			<-time.After(time.Duration(ms) * time.Millisecond)
			delete(MEM, k)
		}()
	}
}

func GET(k string, con net.Conn) {
	// a Type Assertion
	val, ok := MEM[k].(string)
	if !ok {
		con.Write([]byte("$-1\r\n"))

		return
	}

	msg := fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
	con.Write([]byte(msg))
}
