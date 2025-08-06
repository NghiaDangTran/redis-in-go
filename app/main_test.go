package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
)

func sendRedisCommand(conn net.Conn, cmd string) (string, error) {
	_, err := conn.Write([]byte(cmd))
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(conn)

	response, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return response, nil
}

func TestMultipleClients(t *testing.T) {
	var wg sync.WaitGroup

	numClients := 20

	for i := 0; i < numClients; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", "127.0.0.1:6379")
			if err != nil {
				t.Errorf("Client %d failed to connect: %v", id, err)

				return
			}

			defer conn.Close()

			key := fmt.Sprintf("client%d_key", id)
			val := fmt.Sprintf("value%d", id)

			setCmd := fmt.Sprintf(
				"*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
				len(key),
				key,
				len(val),
				val,
			)
			getCmd := fmt.Sprintf(
				"*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n",
				len(key),
				key,
			)

			resp, err := sendRedisCommand(conn, setCmd)
			if err != nil || !strings.HasPrefix(resp, "+OK") {
				t.Errorf("Client %d SET failed: %v (%s)", id, err, resp)
			}

			resp, err = sendRedisCommand(conn, getCmd)
			if err != nil || !strings.Contains(resp, val) {
				t.Errorf("Client %d GET failed: %v (%s)", id, err, resp)
			}

			resp, err = sendRedisCommand(conn, "*1\r\n$4\r\nPING\r\n")
			if err != nil || !strings.HasPrefix(resp, "+PONG") {
				t.Errorf("Client %d PING failed: %v (%s)", id, err, resp)
			}
		}(i)
	}

	wg.Wait()
}
