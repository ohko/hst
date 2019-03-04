package hst

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func handleFunc(hst *HST, method, pattern string, handler ...HandlerFunc) *HST {
	if !hst.DisableRouteLog {
		log.Printf("route:[%s]%s\n", method, pattern)
	}

	f := func(handler HandlerFunc, ctx *Context) {
		start := time.Now()
		defer func() {
			if err := recover(); err != nil {
				switch err.(type) {
				case hstError, *hstError:
					ctx.close = true
				default:
					log.Println(err)
					dep := 0
					for i := 1; i < 10; i++ {
						_, file, line, ok := runtime.Caller(i)
						if !ok {
							break
						}
						if strings.Contains(file, "/runtime/") || strings.Contains(file, "/reflect/") {
							continue
						}
						log.Printf("%s∟%s(%d)\n", strings.Repeat(" ", dep), file, line)
						dep++
					}
					defer func() { recover() }()
					ctx.Data(500, "found error")
				}
			}

			var xforwardfor string
			if ctx.R.Header.Get("Ali-Cdn-Real-Ip") != "" {
				xforwardfor = ctx.R.Header.Get("Ali-Cdn-Real-Ip")
			} else if ctx.R.Header.Get("X-Forwarded-For") != "" {
				xforwardfor = ctx.R.Header.Get("X-Forwarded-For")
			}

			var l io.Writer
			if hst.logger != nil {
				l = hst.logger
			} else {
				l = os.Stdout
			}
			if _, err := l.Write([]byte(logFormatter(&LogData{
				RemoteIP:    strings.Split(ctx.R.RemoteAddr, ":")[0],
				LocalTime:   time.Now(),
				Status:      ctx.status,
				UseTime:     time.Now().Sub(start),
				URI:         fmt.Sprintf("%s %s %s", ctx.R.Method, ctx.R.RequestURI, ctx.R.Proto),
				Sent:        ctx.W.Length(),
				Referer:     ctx.R.Referer(),
				UserAgent:   ctx.R.UserAgent(),
				XForwardFor: xforwardfor,
			}))); err != nil {
				log.Println(err)
			}
		}()

		handler(ctx)
	}

	// 匹配路径
	if _, ok := hst.handleFuncs[pattern]; !ok {
		hst.handleFuncs[pattern] = make(map[string][]HandlerFunc)

		// handler
		hst.handle.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			hs := hst.handleFuncs[pattern]
			for method, hf := range hs {

				// 匹配方法
				if method != "" && method != r.Method {
					continue
				}

				ctx := &Context{hst: hst, W: &responseWriterWithLength{w, 0}, R: r}
				for _, v := range hf {
					f(v, ctx)
					if ctx.close {
						break
					}
				}
			}
		})
	}
	hst.handleFuncs[pattern][method] = handler

	return hst
}

// Shutdown 等待信号，优雅的停止服务
func Shutdown(waitTime time.Duration, hss ...*HST) {
	log.Println("wait ctrl+c ...")

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	for _, v := range hss {
		err := v.shutdown(waitTime)
		if err != nil {
			log.Println(v.s.Addr, err)
		} else {
			log.Println(v.s.Addr, "shutdown ok.")
		}
	}
}

// Request 获取http/https内容
func Request(method, url, cookie, data string, header map[string]string) ([]byte, []*http.Cookie, error) {
	var client *http.Client

	if strings.HasPrefix(url, "https://") {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	req, err := http.NewRequest(method, url, strings.NewReader(data))
	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("Cookie", cookie)
	if header != nil {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	return bs, res.Cookies(), nil
}

// RequestTLS 获取tls内容
func RequestTLS(method, url, ca, crt, key, cookie, data string) ([]byte, []*http.Cookie, error) {
	caCrt, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, nil, err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCrt)

	cliCrt, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		return nil, nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:            pool,
			Certificates:       []tls.Certificate{cliCrt},
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest(method, url, strings.NewReader(data))
	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("Cookie", cookie)
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	return bs, res.Cookies(), nil
}

// MakeGUID 生成唯一的GUID
func MakeGUID() string {
	b := make([]byte, 16)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[8:10], b[6:8], b[4:6], b[10:])
}

/*
MakeTLSFile 生成TLS双向认证证书
# 1.创建根证书密钥文件(自己做CA)root.key：
openssl genrsa -des3 -passout pass:123 -out ssl/root.key 2048
# 2.创建根证书的申请文件root.csr：
openssl req -passin pass:123 -new -subj "/C=CN/ST=Shanghai/L=Shanghai/O=MyCompany/OU=MyCompany/CN=localhost/emailAddress=hk@cdeyun.com" -key ssl/root.key -out ssl/root.csr
# 3.创建根证书root.crt：
openssl x509 -passin pass:123 -req -days 3650 -sha256 -extensions v3_ca -signkey ssl/root.key -in ssl/root.csr -out ssl/root.crt
rm -rf ssl/root.csr

# 1.创建客户端证书私钥
openssl genrsa -des3 -passout pass:456 -out ssl/ssl.key 2048
# 2.去除key口令
openssl rsa -passin pass:456 -in ssl/ssl.key -out ssl/ssl.key
# 3.创建客户端证书申请文件ssl.csr
openssl req -new -subj "/C=CN/ST=Shanghai/L=Shanghai/O=MyCompany/OU=MyCompany/CN=localhost/emailAddress=hk@cdeyun.com" -key ssl/ssl.key -out ssl/ssl.csr
# 4.创建客户端证书文件ssl.crt
openssl x509 -passin pass:123 -req -days 365 -sha256 -extensions v3_req -CA ssl/root.crt -CAkey ssl/root.key -CAcreateserial -in ssl/ssl.csr -out ssl/ssl.crt
rm -rf ssl/ssl.csr
rm -rf ssl/root.srl
# 5.将客户端证书文件ssl.crt和客户端证书密钥文件ssl.key合并成客户端证书安装包ssl.pfx
openssl pkcs12 -export -passout pass:789 -in ssl/ssl.crt -inkey ssl/ssl.key -out ssl/ssl.pfx
*/
func MakeTLSFile(passRoot, passKey, passPfx, path, domain, email string) bool {
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	os.MkdirAll(path, 0755)
	log.Println("Path:", path)
	// 1.1.创建根证书密钥文件(自己做CA)root.key：
	bs, err := exec.Command(`openssl`, `genrsa`, `-des3`, `-passout`, `pass:`+passRoot, `-out`, path+domain+`.ca.key`, `2048`).CombinedOutput()
	log.Println("-", string(bs), err)

	// 1.2.创建根证书的申请文件root.csr：
	exec.Command(`openssl`, `req`, `-passin`, `pass:`+passRoot, `-new`, `-subj`, `/C=CN/ST=Shanghai/L=Shanghai/O=MyCompany/OU=MyCompany/CN=`+domain+`/emailAddress=`+email, `-key`, path+domain+`.ca.key`, `-out`, path+domain+`.ca.csr`).CombinedOutput()

	// 1.3.创建根证书root.crt：
	exec.Command(`openssl`, `x509`, `-passin`, `pass:`+passRoot, `-req`, `-days`, `3650`, `-sha256`, `-extensions`, `v3_ca`, `-signkey`, path+domain+`.ca.key`, `-in`, path+domain+`.ca.csr`, `-out`, path+domain+`.ca.crt`).CombinedOutput()
	exec.Command(`rm`, path+domain+`.ca.csr`).CombinedOutput()

	// 2.1.创建客户端证书私钥
	exec.Command(`openssl`, `genrsa`, `-des3`, `-passout`, `pass:`+passKey, `-out`, path+domain+`.ssl.key`, `2048`).CombinedOutput()

	// 2.2.去除key口令
	exec.Command(`openssl`, `rsa`, `-passin`, `pass:`+passKey, `-in`, path+domain+`.ssl.key`, `-out`, path+domain+`.ssl.key`).CombinedOutput()

	// 2.3.创建客户端证书申请文件ssl.csr
	exec.Command(`openssl`, `req`, `-new`, `-subj`, `/C=CN/ST=Shanghai/L=Shanghai/O=MyCompany/OU=MyCompany/CN=`+domain+`/emailAddress=`+email, `-key`, path+domain+`.ssl.key`, `-out`, path+domain+`.ssl.csr`).CombinedOutput()

	// 2.4.创建客户端证书文件ssl.crt
	exec.Command(`openssl`, `x509`, `-passin`, `pass:`+passRoot, `-req`, `-days`, `365`, `-sha256`, `-extensions`, `v3_req`, `-CA`, path+domain+`.ca.crt`, `-CAkey`, path+domain+`.ca.key`, `-CAcreateserial`, `-in`, path+domain+`.ssl.csr`, `-out`, path+domain+`.ssl.crt`).CombinedOutput()
	exec.Command(`rm`, path+domain+`.ssl.csr`).CombinedOutput()

	// 2.5.将客户端证书文件ssl.crt和客户端证书密钥文件ssl.key合并成客户端证书安装包ssl.pfx
	exec.Command(`openssl`, `pkcs12`, `-export`, `-passout`, `pass:`+passPfx, `-in`, path+domain+`.ssl.crt`, `-inkey`, path+domain+`.ssl.key`, `-out`, path+domain+`.ssl.pfx`).CombinedOutput()
	exec.Command(`rm`, path+domain+`.srl`).CombinedOutput()

	// 3.校验
	bs1, _ := exec.Command(`openssl`, `x509`, `-noout`, `-modulus`, `-in`, path+domain+`.ssl.crt`).CombinedOutput()
	bs2, _ := exec.Command(`openssl`, `rsa`, `-noout`, `-modulus`, `-in`, path+domain+`.ssl.key`).CombinedOutput()
	return string(bs1) == string(bs2)
}
