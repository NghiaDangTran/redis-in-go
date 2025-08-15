// server/server.go
package server

import "sync"

type DataType int

const (
	List DataType = iota
	String
	Stream
)

type Value struct {
	DataType DataType
	Data     any
}

type RedisServer struct {
	MEM   map[string]Value
	CHANS map[string]chan bool
	mu    sync.RWMutex
}

// stream struct
// timemap: map a time to sequece help to find the avaiable seq fast
type StreamData struct {
	TimeMap    map[int]int
	StreamList []StreamEntry
}
type StreamEntry struct {
	Time     int
	Sequence int
	Fields   map[string]string
}

func MEM() map[string]Value {
	if Server == nil {
		panic("Server not initialized")
	}
	return Server.MEM
}

func CHANS() map[string]chan bool {
	if Server == nil {
		panic("Server not initialized")
	}
	return Server.CHANS
}

var Server *RedisServer

func InitServer() {
	Server = &RedisServer{
		MEM:   make(map[string]Value),
		CHANS: make(map[string]chan bool),
	}
}
