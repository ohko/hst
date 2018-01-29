package hst

import (
	"compress/gzip"
	"net/http"
)

// Gzip ...
type Gzip struct {
	gz *gzip.Writer
	rw http.ResponseWriter
}

// NewGzip ...
func NewGzip(w http.ResponseWriter) *Gzip {
	w.Header().Set("Content-Encoding", "gzip")
	gz, _ := gzip.NewWriterLevel(w, gzip.BestCompression)
	return &Gzip{gz: gz, rw: w}
}

func (o *Gzip) Write(bs []byte) (int, error) {
	o.gz.Flush()
	return o.gz.Write(bs)
}

// Header ...
func (o *Gzip) Header() http.Header {
	return o.rw.Header()
}

// WriteHeader ...
func (o *Gzip) WriteHeader(n int) {
	o.rw.WriteHeader(n)
}

// CloseGzip ...
func (o *Gzip) CloseGzip() {
	o.gz.Close()
}
