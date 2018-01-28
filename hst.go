package hst

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

// HST ...
type HST struct {
	s           *http.Server
	handle      *http.ServeMux
	Addr        string
	sessionLock sync.RWMutex
	sessionData map[string]*map[string]*sessionData
}

// HandlerFunc ...
type HandlerFunc func(*Context)

type sessionData struct {
	data   interface{}
	expire time.Time
}

// const
const SESSIONKEY = "HST_SESSION"

// NewHST ...
func NewHST(handler *http.ServeMux) *HST {
	o := new(HST)
	if handler == nil {
		o.handle = http.NewServeMux()
	} else {
		o.handle = handler
	}
	o.sessionData = make(map[string]*map[string]*sessionData)
	return o
}

// ListenHTTP 启动HTTP服务
func (o *HST) ListenHTTP(addr string) error {
	o.s = &http.Server{
		Addr:    addr,
		Handler: o.handle,
	}

	log.Println("Listen http://", addr)
	if err := o.s.ListenAndServe(); err != nil {
		log.Println("Error http://", err)
		return err
	}
	return nil
}

// ListenHTTPS 启动HTTPS服务
func (o *HST) ListenHTTPS(addr, crt, key string) error {
	o.s = &http.Server{
		Addr:    addr,
		Handler: o.handle,
	}

	log.Println("Listen https://", addr)
	if err := o.s.ListenAndServeTLS(crt, key); err != nil {
		log.Println("Error https://", err)
		return err
	}
	return nil
}

// ListenTLS 启动TLS服务
func (o *HST) ListenTLS(addr, ca, crt, key string) error {
	caCrt, err := ioutil.ReadFile(ca)
	if err != nil {
		return err
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

	log.Println("Listen https(tls)://", o.Addr)
	if err := o.s.ListenAndServeTLS(crt, key); err != nil {
		log.Println("Error https(tls)://", err)
		return err
	}
	return nil
}

// HandleFunc ...
// Example:
//		HandleFunc("/", func(c *hst.Context){}, func(c *hst.Context){})
func (o *HST) HandleFunc(pattern string, handler ...HandlerFunc) {
	o.handle.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		c := &Context{
			hst:   o,
			W:     w,
			R:     r,
			close: false,
		}
		for _, v := range handler {
			v(c)
			if c.close {
				break
			}
		}
	})
}

// Shutdown 优雅得关闭服务
func (o *HST) Shutdown(waitTime time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), waitTime)
	defer cancel()
	o.s.Shutdown(ctx)
}

// Favicon 显示favicon.ico
func (o *HST) Favicon() {
	o.handle.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		bs := []byte{0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x10, 0x10, 0x02, 0x00, 0x01, 0x00, 0x01, 0x00, 0xb0, 0x00,
			0x00, 0x00, 0x16, 0x00, 0x00, 0x00, 0x28, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x20, 0x00,
			0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x00, 0x00, 0x00, 0x12, 0x0b,
			0x00, 0x00, 0x12, 0x0b, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x5d, 0x5d,
			0x5d, 0x00, 0xff, 0xff, 0xff, 0x00, 0xff, 0xfb, 0x00, 0x00, 0xff, 0xfb, 0x00, 0x00, 0xff, 0xfb,
			0x00, 0x00, 0xff, 0xfb, 0x00, 0x00, 0xff, 0xe0, 0x00, 0x00, 0xf8, 0x3f, 0x00, 0x00, 0xff, 0xbf,
			0x00, 0x00, 0xf8, 0x3f, 0x00, 0x00, 0xfb, 0xff, 0x00, 0x00, 0xf8, 0x3f, 0x00, 0x00, 0x6f, 0xff,
			0x00, 0x00, 0x6f, 0xff, 0x00, 0x00, 0x6f, 0xff, 0x00, 0x00, 0x0f, 0xff, 0x00, 0x00, 0x6f, 0xff,
			0x00, 0x00, 0x6f, 0xff, 0x00, 0x00, 0xff, 0xfb, 0x00, 0x00, 0xff, 0xfb, 0x00, 0x00, 0xff, 0xfb,
			0x00, 0x00, 0xff, 0xfb, 0x00, 0x00, 0xff, 0xe0, 0x00, 0x00, 0xf8, 0x3f, 0x00, 0x00, 0xff, 0xbf,
			0x00, 0x00, 0xf8, 0x3f, 0x00, 0x00, 0xfb, 0xff, 0x00, 0x00, 0xf8, 0x3f, 0x00, 0x00, 0x6f, 0xff,
			0x00, 0x00, 0x6f, 0xff, 0x00, 0x00, 0x6f, 0xff, 0x00, 0x00, 0x0f, 0xff, 0x00, 0x00, 0x6f, 0xff,
			0x00, 0x00, 0x6f, 0xff, 0x00, 0x00}
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(bs)
	})
}

// Static 静态文件
func (o *HST) Static(partten, path string) {
	o.handle.Handle(partten, http.StripPrefix(partten, http.FileServer(http.Dir(path))))
}

// HandlePfx 输出pfx证书给浏览器安装
// Example:
//		HandlePfx("/ssl.pfx", "/a/b/c.ssl.pfx"))
func (o *HST) HandlePfx(partten, pfxPath string) {
	o.handle.HandleFunc(partten, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-x509-ca-cert")
		caCrt, err := ioutil.ReadFile(pfxPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write(caCrt)
	})
}
