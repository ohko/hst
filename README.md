# hst
**H**ttp/http**S**/**T**ls web服务库

# 功能
- HTTP 普通web服务
- HTTPS 单向认证web服务
- TLS 双向认证web服务
- HTTP认证
- 自签证书生成
- 支持中间件

# 安装
``` shell
go get -v github.com/ohko/hst
```

# 使用
## http
``` golang
h, _ := NewHTTPServer(":8080")
h.HandleFunc("/", func(c *Context) {
    fmt.Fprint(w, "Hello world!")
})
go h.Listen()
```

## https
``` golang
h, _ := NewHTTPSServer(":8081", "ssl.crt", "ssl.key")
h.HandleFunc("/", func(c *Context) {
    fmt.Fprint(w, "Hello world!")
})
go h.Listen()
```

## tls
``` golang
h, _ := NewTLSServer(":8081", "ca.crt", "ssl.crt", "ssl.key")
h.HandleFunc("/", func(c *Context) {
    fmt.Fprint(w, "Hello world!")
})
go h.Listen()
```

# http认证
``` golang
h.HandleFunc("/", BasicAuth("账户", "密码"), func(c *Context) {
    fmt.Fprint(w, "Success")
})
```

# 制作自签证书
``` golang
if !MakeTLSFile("ca证书密码", "https证书密码", "pfx安装证书密码", "证书生成路径", "域名", "邮件地址") {
    t.Fatal("make tls error!")
}
```

# 更多
更多例子查看 [hst_test.go](blob/master/hst_test.go)