# hst
**H**ttp/http**S**/**T**ls web服务库

# 功能
- HTTP 普通web服务
- HTTPS 单向认证web服务
- TLS 双向认证web服务
- 自签证书生成

# 安装
``` shell
go get -v github.com/ohko/hst
```

# 使用
## http
``` golang
h, _ := NewHTTPServer(":8080")
h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Hello world!")
})
go h.Listen()
```

## https
``` golang
h, _ := NewHTTPSServer(":8081", "ssl.crt", "ssl.key")
h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Hello world!")
})
go h.Listen()
```

## tls
``` golang
h, _ := NewTLSServer(":8081", "ca.crt", "ssl.crt", "ssl.key")
h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Hello world!")
})
go h.Listen()
```

# 自签证书
``` golang
if !MakeTLSFile("ca证书密码", "https证书密码", "pfx安装证书密码", "证书生成路径", "域名", "邮件地址") {
    t.Fatal("make tls error!")
}
```