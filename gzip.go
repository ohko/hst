package hst

import (
	"compress/gzip"
	"net/http"
)

type hstGzip struct {
	gz *gzip.Writer
	rw http.ResponseWriter
}

func newGzip(w http.ResponseWriter) *hstGzip {
	w.Header().Set("Content-Encoding", "gzip")
	gz, _ := gzip.NewWriterLevel(w, gzip.BestCompression)
	return &hstGzip{gz: gz, rw: w}
}

func (o *hstGzip) Write(bs []byte) (int, error) {
	o.gz.Flush()
	return o.gz.Write(bs)
}

// Header ...
func (o *hstGzip) Header() http.Header {
	return o.rw.Header()
}

// WriteHeader ...
func (o *hstGzip) WriteHeader(n int) {
	o.rw.WriteHeader(n)
}

// CloseGzip ...
func (o *hstGzip) Close() {
	o.gz.Close()
}
