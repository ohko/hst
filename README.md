# hst
**H**ttp/http**S**/**T**ls web服务库

# 功能
- HTTP 普通web服务
- HTTPS 单向认证web服务
- TLS 双向认证web服务
- HTTP认证
- 自签证书生成
- 支持中间件
- 支持Session
- 支持Render渲染模版
- Static支持gzip输出

# 安装
``` shell
go get -v github.com/ohko/hst
```

# 使用
## http
``` golang
h := New(nil)
h.HandleFunc("/", func(c *Context) {
    fmt.Fprint(w, "Hello world!")
})
h.ListenHTTP(":8080")
```

## https
``` golang
h := New(nil)
h.HandleFunc("/", func(c *Context) {
    fmt.Fprint(w, "Hello world!")
})
go h.ListenHTTPS(":8081", "ssl.crt", "ssl.key")
log.Println("wait ctrl+c ...")
Shutdown(time.Second*5, h)
```

## tls
``` golang
h := NewTLSServer(nil)
h.HandleFunc("/", func(c *Context) {
    fmt.Fprint(w, "Hello world!")
})
go h.ListenTLS(":8081", "ca.crt", "ssl.crt", "ssl.key")
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

# TODO
- 日志
- 分开GET,POST,PUT,PATCH,DELETE,OPTIONS
- 路径参数
- 表单验证
- 上传文件，分片上传
- JSONP
- 静态文件打包
- http2 push
- Cookie
- 多服务器
- 自定义HTTP配置

# 更多
更多例子查看 [hst_test.go](blob/master/hst_test.go)