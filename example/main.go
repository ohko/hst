package main

import (
	"fmt"
	"html/template"
	"time"

	"github.com/ohko/hst"
)

func main() {
	s := hst.New(nil)

	// HTML模版
	s.SetDelims("{[{", "}]}")
	s.SetTemplateFunc(template.FuncMap{"json": func(x string) string { return "JSON:" + x }})
	// s.ParseGlob("./template/**/*")
	s.ParseFiles("./template/index/index.html", "./template/sub/sub.html")

	// 自动路由
	s.RegisterHandle(&Auto{})

	// 自定义路由
	s.HandleFunc("/", func(c *hst.Context) {
		c.HTML(200, "index/index.html", "from index")
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

	// s.ListenAutoCert(".https", "xx.example.com")
	s.ListenHTTP(":8080")
}
