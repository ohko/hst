package hst

import (
	"log"
	"net/http"
)

// HTTPServer http服务
type HTTPServer struct {
	base
}

// NewHTTPServer ...
func NewHTTPServer(addr string) (HST, error) {
	o := new(HTTPServer)
	o.Addr = addr
	o.handle = http.NewServeMux()
	o.s = &http.Server{
		Addr:    addr,
		Handler: o.handle,
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
