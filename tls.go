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
	base
	Ca  string
	Crt string
	Key string
}

// NewTLSServer ...
func NewTLSServer(addr, ca, crt, key string) (HST, error) {
	o := new(TLSServer)
	o.Addr = addr
	o.Ca = ca
	o.Crt = crt
	o.Key = key
	o.handle = http.NewServeMux()

	caCrt, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCrt)
	o.s = &http.Server{
		Addr:    addr,
		Handler: o.handle,
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
