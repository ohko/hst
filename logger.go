package hst

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"
)

// LogData 127.0.0.1 [2006-01-02 15:04:05] 200 123s "GET / HTTP/1.1" 1234 "referer" "user_agent" "http_x_forwarded_for"
type LogData struct {
	RemoteIP    string
	LocalTime   time.Time
	Status      int
	UseTime     time.Duration
	URI         string
	Sent        int
	Referer     string
	UserAgent   string
	XForwardFor string
}

var logFormatter = func(p *LogData) string {
	return fmt.Sprintf(`%s [%s] %d "%s" %d "%s" "%s" "%s" %v`,
		p.RemoteIP, p.LocalTime.Format("2006-01-02 15:04:05"), p.Status,
		p.URI, p.Sent, p.Referer, p.UserAgent, p.XForwardFor, p.UseTime,
	) + "\n"
}

type responseWriterWithLength struct {
	http.ResponseWriter
	length int
}

func (w *responseWriterWithLength) Write(b []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(b)

	w.length += n

	return
}

func (w *responseWriterWithLength) Length() int {
	return w.length
}

func (w *responseWriterWithLength) Hijack() (conn net.Conn, rw *bufio.ReadWriter, err error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}
