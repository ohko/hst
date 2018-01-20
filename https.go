package hst

import (
	"log"
	"net/http"
)

// HTTPSServer 单向验证
type HTTPSServer struct {
	base
	Crt string
	Key string
}

// NewHTTPSServer ...
func NewHTTPSServer(addr, crt, key string) (HST, error) {
	o := new(HTTPSServer)
	o.Addr = addr
	o.Crt = crt
	o.Key = key
	o.handle = http.NewServeMux()
	o.s = &http.Server{
		Addr:    addr,
		Handler: o.handle,
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
