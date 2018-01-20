package hst

import (
	"log"
	"net/http"
)

// HTTPSServer 单向验证
type HTTPSServer struct {
	s      *http.Server
	Addr   string
	Crt    string
	Key    string
	Handle *http.ServeMux
}

// NewHTTPSServer ...
func NewHTTPSServer(addr, crt, key string) (HST, error) {
	o := &HTTPSServer{
		Addr:   addr,
		Crt:    crt,
		Key:    key,
		Handle: http.NewServeMux(),
	}
	o.s = &http.Server{
		Addr:    addr,
		Handler: o.Handle,
	}
	return o, nil
}

// Listen ...
func (o *HTTPSServer) Listen() error {
	log.Println("Listen https://", o.Addr)
	if err := o.s.ListenAndServeTLS(o.Crt, o.Key); err != nil {
		return err
	}
	return nil
}

// HandleFunc ...
func (o *HTTPSServer) HandleFunc(pattern string, handler http.HandlerFunc) {
	o.Handle.HandleFunc(pattern, handler)
}
