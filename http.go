package hst

import (
	"log"
	"net/http"
)

// HTTPServer http服务
type HTTPServer struct {
	s      *http.Server
	Addr   string
	Handle *http.ServeMux
}

// NewHTTPServer ...
func NewHTTPServer(addr string) (HST, error) {
	o := &HTTPServer{
		Addr:   addr,
		Handle: http.NewServeMux(),
	}
	o.s = &http.Server{
		Addr:    addr,
		Handler: o.Handle,
	}
	return o, nil
}

// Listen ...
func (o *HTTPServer) Listen() error {
	log.Println("Listen http://", o.Addr)
	if err := o.s.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

// HandleFunc ...
func (o *HTTPServer) HandleFunc(pattern string, handler http.HandlerFunc) {
	o.Handle.HandleFunc(pattern, handler)
}
