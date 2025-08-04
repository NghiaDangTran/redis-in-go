package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

var _ = net.Listen
var _ = os.Exit

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

		cmd := strings.Split(string(b[:numBytes]), "\r\n")[2]

		fmt.Printf("User Command: \"%s\" \n", cmd)

		switch strings.ToUpper(cmd) {
		case "PING":
			con.Write([]byte("+PONG\r\n"))
		case "QUIT":
			return
		default:
			con.Write([]byte("-ERR unknown command\r\n"))
		}

	}

}
