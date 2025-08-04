package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var _ = net.Listen
var _ = os.Exit

// avalaibel around 10000 key
var MEM = make(map[string]string, 10000)

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

		fmt.Printf("User Command: \"%s\" \n", cmd)

		switch strings.ToUpper(cmd[2]) {
		case "PING":
			con.Write([]byte("+PONG\r\n"))
		case "ECHO":
			if len(cmd) < 5 {
				con.Write([]byte("-ERR Not enough arguments for ECHO command \r\n"))
			} else {
				message := cmd[4]
				con.Write([]byte(fmt.Sprintf("+%s\r\n", message)))

			}

		case "QUIT":
			return
		case "SET":
			// sample "[*3 $3 set $4 test $2 ok ]"
			if len(cmd) < 6 {
				con.Write([]byte("-ERR Not enough arguments for SET command \r\n"))
			} else {
				key := cmd[4]
				value := cmd[6]
				// with expiry ags  "[*5 $3 SET $3 foo $3 bar $2 px $3 100 ]"
				SET(key, value, con, cmd...)

			}

		case "GET":
			// sample: "[*2 $3 get $2 hi ]"
			if len(cmd) < 4 {
				con.Write([]byte("-ERR Not enough arguments for GET command \r\n"))
			} else {
				key := cmd[4]
				GET(key, con)
			}

		default:
			con.Write([]byte("-ERR unknown command\r\n"))
		}

	}

}

func SET(k string, v string, con net.Conn, agr ...string) {
	// key already exist in memory
	// _, ok := MEM[k]
	// if ok {
	// 	con.Write([]byte("-ERR This Key Already Existed\r\n"))
	// 	return
	// }
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
	val, ok := MEM[k]
	if !ok {
		con.Write([]byte("$-1\r\n"))
		return
	}
	msg := fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
	con.Write([]byte(msg))

}
