package hst

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
)

// TLSServer 双向验证
type TLSServer struct {
	s      *http.Server
	Addr   string
	Ca     string
	Crt    string
	Key    string
	Handle *http.ServeMux
}

// NewTLSServer ...
func NewTLSServer(addr, ca, crt, key string) (HST, error) {
	o := &TLSServer{
		Addr:   addr,
		Ca:     ca,
		Crt:    crt,
		Key:    key,
		Handle: http.NewServeMux(),
	}

	caCrt, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCrt)
	o.s = &http.Server{
		Addr:    addr,
		Handler: o.Handle,
		TLSConfig: &tls.Config{
			ClientCAs:  pool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
	}
	return o, nil
}

// Listen ...
func (o *TLSServer) Listen() error {
	log.Println("Listen tls://", o.Addr)
	if err := o.s.ListenAndServeTLS(o.Crt, o.Key); err != nil {
		return err
	}
	return nil
}

// HandleFunc ...
func (o *TLSServer) HandleFunc(pattern string, handler http.HandlerFunc) {
	o.Handle.HandleFunc(pattern, handler)
}
