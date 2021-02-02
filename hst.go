package hst

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

// HST ...
type HST struct {
	s               *http.Server
	handle          *http.ServeMux
	hs              *Handlers
	Addr            string
	session         Session
	CrossOrigin     string // 支持跨域 "*" / "a.com,b.com"
	DisableRouteLog bool   // 禁止显示启动时的route路径显示
	ContentSecurityPolicyReportOnly string // https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Content-Security-Policy-Report-Only

	DisableAutoGzip bool // 禁止自动gzip

	// template 读取模版到内存
	template        *template.Template
	templateDelims  *delims
	templateFuncMap template.FuncMap
	// 用于template中找不到时，去模版目录读取模版文件
	templatePath string

	handleFuncs map[string]map[string][]HandlerFunc
	logger      io.Writer
}

// HandlerFunc ...
type HandlerFunc func(*Context)

// hstError 用于提前终止流程
type hstError struct{ s string }

func (o *hstError) Error() string { return o.s }

type delims struct {
	left  string
	right string
}

// New ...
func New(handlers *Handlers) *HST {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	runtime.GOMAXPROCS(runtime.NumCPU())
	o := new(HST)
	o.handle = http.NewServeMux()
	o.hs = handlers
	o.templatePath = "./"
	o.templateDelims = &delims{left: "{{", right: "}}"}
	o.handleFuncs = make(map[string]map[string][]HandlerFunc)
	return o
}

// ListenAutoCert 同时监听http/https，自动获取https证书
func (o *HST) ListenAutoCert(cacheDir string, hosts ...string) error {
	m := &autocert.Manager{
		Cache:      autocert.DirCache(cacheDir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(hosts...),
	}

	log.Println("Listen http://", hosts)
	go func() {
		log.Println(http.ListenAndServe(":http", m.HTTPHandler(nil)))
	}()

	o.s = &http.Server{
		Addr:      ":https",
		Handler:   o.handle,
		TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
	}
	if o.hs != nil {
		for k, v := range *o.hs {
			o.HandleFunc(k, v...)
		}
	}

	log.Println("Listen https://", hosts)
	if err := o.s.ListenAndServeTLS("", ""); err != nil {
		log.Println("Error https://", hosts, err)
		return err
	}
	return nil
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
		log.Println("Error http://", addr, err)
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
		log.Println("Error https://", addr, err)
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
		log.Println("Error https(tls)://", o.Addr, err)
		return err
	}
	return nil
}

// Handle 添加路由
// Example:
//		Handle("/", http.Handler)
func (o *HST) Handle(pattern string, handler http.Handler) *HST {
	o.handle.Handle(pattern, handler)
	return o
}

// HandleFunc 添加路由
// Example:
//		HandleFunc("/", func(c *hst.Context){}, func(c *hst.Context){})
func (o *HST) HandleFunc(pattern string, handler ...HandlerFunc) *HST {
	return handleFunc(o, "", pattern, handler...)
}

// GET ...
func (o *HST) GET(pattern string, handler ...HandlerFunc) *HST {
	return handleFunc(o, "GET", pattern, handler...)
}

// POST ...
func (o *HST) POST(pattern string, handler ...HandlerFunc) *HST {
	return handleFunc(o, "POST", pattern, handler...)
}

// PUT ...
func (o *HST) PUT(pattern string, handler ...HandlerFunc) *HST {
	return handleFunc(o, "PUT", pattern, handler...)
}

// PATCH ...
func (o *HST) PATCH(pattern string, handler ...HandlerFunc) *HST {
	return handleFunc(o, "PATCH", pattern, handler...)
}

// DELETE ...
func (o *HST) DELETE(pattern string, handler ...HandlerFunc) *HST {
	return handleFunc(o, "DELETE", pattern, handler...)
}

// OPTIONS ...
func (o *HST) OPTIONS(pattern string, handler ...HandlerFunc) *HST {
	return handleFunc(o, "OPTIONS", pattern, handler...)
}

// RegisterHandle 注册自动路由
// Example:
//		RegisterHandle(&User{}, &Other{})
func (o *HST) RegisterHandle(middleware []HandlerFunc, classes ...interface{}) *HST {
	fixName := func(name string) string {
		r := []rune(name)
		a := map[rune]rune{'A': 'a', 'B': 'b', 'C': 'c', 'D': 'd', 'E': 'e', 'F': 'f', 'G': 'g', 'H': 'h', 'I': 'i', 'J': 'j', 'K': 'k', 'L': 'l', 'M': 'm', 'N': 'n', 'O': 'o', 'P': 'p', 'Q': 'q', 'R': 'r', 'S': 's', 'T': 't', 'U': 'u', 'V': 'v', 'W': 'w', 'X': 'x', 'Y': 'y', 'Z': 'z'}
		b := map[string]string{"A": "_a", "B": "_b", "C": "_c", "D": "_d", "E": "_e", "F": "_f", "G": "_g", "H": "_h", "I": "_i", "J": "_j", "K": "_k", "L": "_l", "M": "_m", "N": "_n", "O": "_o", "P": "_p", "Q": "_q", "R": "_r", "S": "_s", "T": "_t", "U": "_u", "V": "_v", "W": "_w", "X": "_x", "Y": "_y", "Z": "_z"}

		// 首字母小写
		if v, ok := a[r[0]]; ok {
			r[0] = v
		}

		// 除首字母外，其它大写字母替换成下划线加小写
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
			o.HandleFunc(path, append(middleware, func(v reflect.Value) HandlerFunc {
				return func(c *Context) { v.Call([]reflect.Value{reflect.ValueOf(c)}) }
			}(reflect.ValueOf(c).Method(i)))...)
		}
	}
	return o
}

// Shutdown 优雅得关闭服务
func (o *HST) shutdown(waitTime time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), waitTime)
	defer cancel()
	return o.s.Shutdown(ctx)
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
	return o.Handle(partten, http.StripPrefix(partten, http.FileServer(http.Dir(path))))
}

// StaticGzip 静态文件，增加gzip压缩
func (o *HST) StaticGzip(partten, path string) *HST {
	o.handle.HandleFunc(partten, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		if o.ContentSecurityPolicyReportOnly != "" {
			w.Header().Set("Content-Security-Policy-Report-Only", o.ContentSecurityPolicyReportOnly)
		}
		gz := newGzip(w)
		http.StripPrefix(partten, http.FileServer(http.Dir(path))).ServeHTTP(gz, r)
		gz.Close()
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
	o.templateDelims = &delims{left: left, right: right}
	return o
}

// SetTemplateFunc 设置模板函数
func (o *HST) SetTemplateFunc(funcMap template.FuncMap) *HST {
	o.templateFuncMap = funcMap
	return o
}

// SetSession 设置Session
func (o *HST) SetSession(sess Session) *HST {
	o.session = sess
	return o
}

// ParseGlob 预加载模版文件
func (o *HST) ParseGlob(pattern string) *HST {
	o.template = template.Must(template.New("").
		Delims(o.templateDelims.left, o.templateDelims.right).
		Funcs(o.templateFuncMap).
		ParseGlob(pattern))
	return o
}

// ParseFiles 预加载模版文件
func (o *HST) ParseFiles(filenames ...string) *HST {
	o.template = template.Must(template.New("").
		Delims(o.templateDelims.left, o.templateDelims.right).
		Funcs(o.templateFuncMap).
		ParseFiles(filenames...))
	return o
}

// SetTemplatePath 设置模版文件跟路径
func (o *HST) SetTemplatePath(pathname string) *HST {
	if !strings.HasSuffix(pathname, "/") {
		pathname += "/"
	}
	o.templatePath = pathname
	return o
}

// SetLogger 设置日志记录
func (o *HST) SetLogger(logger io.Writer) *HST {
	o.logger = logger
	return o
}
