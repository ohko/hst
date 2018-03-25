package hst

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// HST ...
type HST struct {
	s       *http.Server
	handle  *http.ServeMux
	hs      *Handlers
	Addr    string
	session Session

	// template
	templateDelims []string
	templatePath   string
	layout         map[string][]string
}

// HandlerFunc ...
type HandlerFunc func(*Context)

// H ...
type H map[string]interface{}

// hstError 用于提前终止流程
type hstError struct{ s string }

func (o *hstError) Error() string { return o.s }

// NewHST ...
func NewHST(handlers *Handlers) *HST {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	runtime.GOMAXPROCS(runtime.NumCPU())
	o := new(HST)
	o.session = NewMemorySession()
	o.handle = http.NewServeMux()
	o.hs = handlers
	o.layout = make(map[string][]string)
	return o
}

// ListenHTTP 启动HTTP服务
func (o *HST) ListenHTTP(addr string) error {
	o.s = &http.Server{
		Addr:    addr,
		Handler: o.handle,
	}
	if o.hs != nil {
		for k, v := range *o.hs {
			o.HandleFunc(k, v...)
		}
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
	if o.hs != nil {
		for k, v := range *o.hs {
			o.HandleFunc(k, v...)
		}
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
	if o.hs != nil {
		for k, v := range *o.hs {
			o.HandleFunc(k, v...)
		}
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
func (o *HST) HandleFunc(pattern string, handler ...HandlerFunc) *HST {
	log.Println("handle:", pattern)
	o.handle.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		c := &Context{
			hst:     o,
			session: o.session,
			W:       w,
			R:       r,
			close:   false,
		}
		for _, v := range handler {
			func(v HandlerFunc, c *Context) {
				defer func() {
					if err := recover(); err != nil {
						switch err.(type) {
						case hstError, *hstError:
							c.close = true
						default:
							log.Println(err)
							c.HTML("found error")
						}
					}
				}()
				v(c)
			}(v, c)
			if c.close {
				break
			}
		}
	})
	return o
}

// RegisterHandle ...
// Example:
//		RegisterHandle(&User{}, &Other{})
func (o *HST) RegisterHandle(classes ...interface{}) *HST {
	fixName := func(name string) string {
		r := []rune(name)
		a := map[rune]rune{'A': 'a', 'B': 'b', 'C': 'c', 'D': 'd', 'E': 'e', 'F': 'f', 'G': 'g', 'H': 'h', 'I': 'i', 'J': 'j', 'K': 'k', 'L': 'l', 'M': 'm', 'N': 'n', 'O': 'o', 'P': 'p', 'Q': 'q', 'R': 'r', 'S': 's', 'T': 't', 'U': 'u', 'V': 'v', 'W': 'w', 'X': 'x', 'Y': 'y', 'Z': 'z'}
		b := map[string]string{"A": "_a", "B": "_b", "C": "_c", "D": "_d", "E": "_e", "F": "_f", "G": "_g", "H": "_h", "I": "_i", "J": "_j", "K": "_k", "L": "_l", "M": "_m", "N": "_n", "O": "_o", "P": "_p", "Q": "_q", "R": "_r", "S": "_s", "T": "_t", "U": "_u", "V": "_v", "W": "_w", "X": "_x", "Y": "_y", "Z": "_z"}

		if v, ok := a[r[0]]; ok {
			r[0] = v
		}

		s := string(r)
		for k, v := range b {
			s = strings.Replace(s, k, v, -1)
		}
		return s
	}

	for _, c := range classes {
		name := reflect.TypeOf(c).Elem().Name()
		if strings.HasSuffix(name, "Controller") {
			name = name[:len(name)-10]
		}
		name = "/" + fixName(name)
		if name == "/index" {
			name = ""
		}
		for i := 0; i < reflect.TypeOf(c).NumMethod(); i++ {
			method := "/" + fixName(reflect.TypeOf(c).Method(i).Name)
			if method == "/index" {
				method = "/"
			}
			path := name + method
			o.HandleFunc(path, func(v reflect.Value) HandlerFunc {
				return func(c *Context) { v.Call([]reflect.Value{reflect.ValueOf(c)}) }
			}(reflect.ValueOf(c).Method(i)))
		}
	}
	return o
}

// Shutdown 优雅得关闭服务
func (o *HST) Shutdown(waitTime time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), waitTime)
	defer cancel()
	o.s.Shutdown(ctx)
}

// Favicon 显示favicon.ico
func (o *HST) Favicon() *HST {
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
	return o
}

// Static 静态文件
func (o *HST) Static(partten, path string) *HST {
	o.handle.Handle(partten, http.StripPrefix(partten, http.FileServer(http.Dir(path))))
	return o
}

// StaticGzip 静态文件，增加gzip压缩
func (o *HST) StaticGzip(partten, path string) *HST {
	o.handle.HandleFunc(partten, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := NewGzip(w)
		http.StripPrefix(partten, http.FileServer(http.Dir(path))).ServeHTTP(gz, r)
		gz.CloseGzip()
	})
	return o
}

// HandlePfx 输出pfx证书给浏览器安装
// Example:
//		HandlePfx("/ssl.pfx", "/a/b/c.ssl.pfx"))
func (o *HST) HandlePfx(partten, pfxPath string) *HST {
	o.handle.HandleFunc(partten, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-x509-ca-cert")
		caCrt, err := ioutil.ReadFile(pfxPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write(caCrt)
	})
	return o
}

// SetDelims 定义模板符号
func (o *HST) SetDelims(left, right string) *HST {
	o.templateDelims = []string{left, right}
	return o
}

// SetTemplatePath 设置模板路径
func (o *HST) SetTemplatePath(path string) *HST {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	o.templatePath = path
	return o
}

// SetLayout 定义layout模板
func (o *HST) SetLayout(name string, files ...string) *HST {
	o.layout[name] = files
	return o
}
