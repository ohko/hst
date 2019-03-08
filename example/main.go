package main

import (
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/ohko/hst"
)

func main() {
	s := hst.New(nil)

	// 禁止显示Route日志
	// s.DisableRouteLog = true

	// HTML模版
	s.SetDelims("{[{", "}]}")
	s.SetTemplateFunc(template.FuncMap{"json": func(x string) string { return "JSON:" + x }})
	// s.ParseGlob("./template/**/*")
	s.ParseFiles("./template/index/index.html", "./template/sub/sub.html")
	s.SetTemplatePath("./template/")

	// 自动路由
	s.RegisterHandle([]hst.HandlerFunc{func(c *hst.Context) {
		// 检查用户权限
		// ...
	}}, &Auto{})

	// 自定义路由
	s.HandleFunc("/", func(c *hst.Context) {
		c.SetCookie("cn", "cv", 3600, "/", "", false, true)
		c.HTML(200, "index/index.html", "from index")
	})
	s.HandleFunc("/refresh", func(c *hst.Context) {
		c.HTML2(200, "index/index.html", "from index", "sub/sub.html")
	})
	s.HandleFunc("/404", func(c *hst.Context) {
		panic(404)
	})
	s.HandleFunc("/ip", func(c *hst.Context) {
		fmt.Fprintln(c.W, c.R.RemoteAddr)
	})
	s.HandleFunc("/time", func(c *hst.Context) {
		loc, _ := time.LoadLocation("Asia/Shanghai")
		fmt.Fprintln(c.W, time.Now().In(loc).Format("2006-01-02 15:04:05"))
	})

	// 路由分组
	s.Group("/group1", func(c *hst.Context) {
		c.JSON2(200, 0, "group1")
	})
	g := s.Group("/group", nil)
	g.HandleFunc("/sub1", func(c *hst.Context) {
		c.HTML(200, "sub/sub.html", "hello sub1")
	})
	g.HandleFunc("/sub2", func(c *hst.Context) {
		fmt.Fprintln(c.W, "hello sub2")
	})

	// GET/POST
	g.GET("/get", func(ctx *hst.Context) { ctx.Data(200, "get") })
	g.GET("/getpost", func(ctx *hst.Context) { ctx.Data(200, "getpost:get") })
	g.POST("/getpost", func(ctx *hst.Context) { ctx.Data(200, "getpost:post") })

	// logger
	l, _ := os.OpenFile("/tmp/hst.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	s.SetLogger(l)

	// 静态文件
	s.StaticGzip("/static/", "./template/")

	// ajax
	s.HandleFunc("/ajax", func(ctx *hst.Context) {
		str := strings.Repeat(".", 2048)
		ctx.JSON2(200, 0, str)
	})

	// s.ListenAutoCert(".https", "xx.example.com")
	s.ListenHTTP(":8080")
}
