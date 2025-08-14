package commands

import (
	"bytes"
	"net"
	"testing"
	"time"

	srv "github.com/codecrafters-io/redis-starter-go/server"
)

// testConn is a minimal net.Conn that captures writes.
type testConn struct{ buf bytes.Buffer }

func (c *testConn) Read(b []byte) (int, error)         { return 0, nil }
func (c *testConn) Write(b []byte) (int, error)        { return c.buf.Write(b) }
func (c *testConn) Close() error                       { return nil }
func (c *testConn) LocalAddr() net.Addr                { return dummyAddr("local") }
func (c *testConn) RemoteAddr() net.Addr               { return dummyAddr("remote") }
func (c *testConn) SetDeadline(t time.Time) error      { return nil }
func (c *testConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *testConn) SetWriteDeadline(t time.Time) error { return nil }

type dummyAddr string

func (a dummyAddr) Network() string { return string(a) }
func (a dummyAddr) String() string  { return string(a) }

func newConn() *testConn { return &testConn{} }

func TestPing(t *testing.T) {
	srv.InitServer()
	con := newConn()
	Ping(con)
	if got := con.buf.String(); got != "+PONG\r\n" {
		t.Fatalf("unexpected response: %q", got)
	}
}

func TestEcho(t *testing.T) {
	srv.InitServer()
	con := newConn()
	Echo("hello", con)
	if got := con.buf.String(); got != "+hello\r\n" {
		t.Fatalf("unexpected response: %q", got)
	}
}

func TestSetAndGet(t *testing.T) {
	srv.InitServer()
	con := newConn()
	Set("k", "v", con)
	if got := con.buf.String(); got != "+OK\r\n" {
		t.Fatalf("unexpected SET response: %q", got)
	}
	// GET returns bulk string
	con2 := newConn()
	Get("k", con2)
	if got := con2.buf.String(); got != "$1\r\nv\r\n" {
		t.Fatalf("unexpected GET response: %q", got)
	}
}

func TestSetPXExpiry(t *testing.T) {
	srv.InitServer()
	con := newConn()
	Set("exp", "1", con, "PX", "10")
	time.Sleep(20 * time.Millisecond)
	con2 := newConn()
	Get("exp", con2)
	if got := con2.buf.String(); got != "$-1\r\n" {
		t.Fatalf("expected expired key, got: %q", got)
	}
}

func TestTypeOf(t *testing.T) {
	srv.InitServer()
	con := newConn()
	Type("missing", con)
	if got := con.buf.String(); got != "+none\r\n" {
		t.Fatalf("unexpected TYPE for missing: %q", got)
	}
	// string
	con.buf.Reset()
	Set("s", "v", newConn())
	Type("s", con)
	if got := con.buf.String(); got != "+string\r\n" {
		t.Fatalf("unexpected TYPE for string: %q", got)
	}
	// list
	con.buf.Reset()
	LPush("l", []string{"$1", "a"}, newConn())
	Type("l", con)
	if got := con.buf.String(); got != "+list\r\n" {
		t.Fatalf("unexpected TYPE for list: %q", got)
	}
}

func TestLPushLLenAndLRange(t *testing.T) {
	srv.InitServer()
	con := newConn()
	// LPUSH list a b => values passed as ["$1","a","$1","b"]
	LPush("list", []string{"$1", "a", "$1", "b"}, con)
	if got := con.buf.String(); got != ":2\r\n" {
		t.Fatalf("unexpected LPUSH reply: %q", got)
	}
	// LLEN
	con.buf.Reset()
	LLen("list", con)
	if got := con.buf.String(); got != ":2\r\n" {
		t.Fatalf("unexpected LLEN: %q", got)
	}
	// LRANGE 0 -1 -> b, a (LPUSH prepends in reverse of input)
	con.buf.Reset()
	LRange("list", 0, -1, con)
	if got := con.buf.String(); got != "*2\r\n$1\r\nb\r\n$1\r\na\r\n" {
		t.Fatalf("unexpected LRANGE: %q", got)
	}
}

func TestRPushThenLPopVariants(t *testing.T) {
	srv.InitServer()
	// RPUSH list a b c => values passed as ["$1","a","$1","b","$1","c"]
	RPush("list", []string{"$1", "a", "$1", "b", "$1", "c"}, newConn())

	con := newConn()
	// LPOP single
	LPop("list", 0, con)
	if got := con.buf.String(); got != "$1\r\na\r\n" {
		t.Fatalf("unexpected LPOP single: %q", got)
	}
	// LPOP count=2 -> returns array of next two
	con.buf.Reset()
	LPop("list", 2, con)
	if got := con.buf.String(); got != "*2\r\n$1\r\nb\r\n$1\r\nc\r\n" {
		t.Fatalf("unexpected LPOP multi: %q", got)
	}
	// further LPOP -> nil bulk
	con.buf.Reset()
	LPop("list", 0, con)
	if got := con.buf.String(); got != "$-1\r\n" {
		t.Fatalf("unexpected LPOP empty: %q", got)
	}
}

func TestLRangeNegativeIndices(t *testing.T) {
	srv.InitServer()
	RPush("list", []string{"$1", "a", "$1", "b", "$1", "c"}, newConn())
	con := newConn()
	LRange("list", -2, -1, con)
	if got := con.buf.String(); got != "*2\r\n$1\r\nb\r\n$1\r\nc\r\n" {
		t.Fatalf("unexpected LRANGE -2 -1: %q", got)
	}
}

func TestBLPopTimeout(t *testing.T) {
	srv.InitServer()
	con := newConn()
	done := make(chan string, 1)
	go func() {
		BLPop("list", 30*time.Millisecond, con)
		done <- con.buf.String()
	}()
	select {
	case out := <-done:
		if out != "$-1\r\n" {
			t.Fatalf("unexpected BLPOP timeout output: %q", out)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("BLPOP timeout test did not finish in time")
	}
}

func TestBLPopSuccessAfterRPush(t *testing.T) {
	srv.InitServer()
	con := newConn()
	done := make(chan string, 1)
	go func() {
		BLPop("list", 500*time.Millisecond, con)
		done <- con.buf.String()
	}()
	// Give BLPOP a moment to register its channel
	time.Sleep(10 * time.Millisecond)
	// Now RPUSH to trigger notification
	RPush("list", []string{"$1", "x"}, newConn())
	select {
	case out := <-done:
		expected := "*2\r\n$4\r\nlist\r\n$1\r\nx\r\n"
		if out != expected {
			t.Fatalf("unexpected BLPOP output: %q", out)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("BLPOP did not complete after RPUSH")
	}
}
