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
	"strings"
	"time"
)

// Shutdown 等待信号，优雅的停止服务
func Shutdown(hs []*HST, waitTime time.Duration, sig ...os.Signal) {
	c := make(chan os.Signal)
	signal.Notify(c, sig...)
	<-c

	for _, v := range hs {
		v.Shutdown(waitTime)
	}
}

// HTTPRequest 获取http内容
func HTTPRequest(method, url, cookie, data string) ([]byte, []*http.Cookie, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(data))
	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("Cookie", cookie)
	res, err := (&http.Client{}).Do(req)
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

// HTTPSRequest 获取https内容
func HTTPSRequest(method, url, cookie, data string) ([]byte, []*http.Cookie, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
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
	defer res.Body.Close()
	bs, err := ioutil.ReadAll(res.Body)
	return bs, res.Cookies(), nil
}

// TLSSRequest 获取tls内容
func TLSSRequest(method, url, ca, crt, key, cookie, data string) ([]byte, []*http.Cookie, error) {
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
	defer res.Body.Close()
	bs, err := ioutil.ReadAll(res.Body)
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
