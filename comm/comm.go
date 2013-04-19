package comm

import "bufio"
import "fmt"
import "log"
import "strings"
import "net"
import "flag"

var addr = flag.String("addr", "", "")
var username = flag.String("user", "", "")
var password = flag.String("pass", "", "")
var traceNet = flag.Bool("trace_net", false, "")

type Conn struct {
	conn *net.TCPConn
	rd   *bufio.Reader
}

func (c *Conn) ReadRawLine() (string, error) {
	s, err := c.rd.ReadString('\n')
	if err != nil {
		log.Printf("Network read error: %s", err.Error())
	} else if *traceNet {
		log.Printf("<- %s", s)
	}
	if s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	return s, err
}

func (c *Conn) Printf(format string, a ...interface{}) error {
	s := fmt.Sprintf(format, a...)
	if *traceNet {
		log.Printf("-> %s", s)
	}
	_, err := fmt.Fprintf(c.conn, "%s\n", s)
	if err != nil {
		log.Printf("Network write error: %s", err.Error())
	}
	return err
}

func (c *Conn) expectLine(exp string) error {
	s, err := c.ReadRawLine()
	if err != nil {
		return err
	}
	if s != exp {
		return fmt.Errorf("Expected [%s], got [%s]", exp, s)
	}
	return nil
}

type RemoteError struct {
	Code int
	Msg  string
}

func (r RemoteError) Error() string {
	return fmt.Sprintf("Remote error %d (%s)", r.Code, r.Msg)
}

func (c *Conn) ReadResult() error {
	s, err := c.ReadRawLine()
	if err != nil {
		return err
	}
	if s == "OK" {
		return nil
	}
	ss := strings.SplitN(s, " ", 3)
	if len(ss) != 3 || ss[0] != "FAILED" {
		return fmt.Errorf("Wrong result: %s", s)
	}
	re := RemoteError{Msg: ss[2]}
	fmt.Sscanf(ss[1], "%d", &re.Code)
	return re
}

func NewConn() (*Conn, error) {
	c := new(Conn)
	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		return nil, err
	}
	c.conn = conn.(*net.TCPConn)
	c.conn.SetNoDelay(true)
	c.rd = bufio.NewReader(c.conn)
	// login:
	if err := c.expectLine("LOGIN"); err != nil {
		return nil, err
	}
	if err := c.Printf(*username); err != nil {
		return nil, err
	}
	if err = c.expectLine("PASS"); err != nil {
		return nil, err
	}
	if err := c.Printf(*password); err != nil {
		return nil, err
	}
	if err := c.ReadResult(); err != nil {
		return nil, err
	}
	return c, nil
}
