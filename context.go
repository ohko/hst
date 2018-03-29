package hst

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"time"
)

// Context 上下文数据
type Context struct {
	hst     *HST
	session Session
	W       http.ResponseWriter
	R       *http.Request
	close   bool

	// template
	templateFuncMap template.FuncMap
}

// Close 结束后面的流程
func (o *Context) Close() {
	o.close = true
	panic(&hstError{"end"})
}

// JSON 返回json数据，自动识别jsonp
func (o *Context) JSON(data interface{}) error {
	defer o.Close()
	if o.hst.CrossOrigin != "" {
		crossOrigin := o.hst.CrossOrigin
		if o.hst.CrossOrigin == "*" {
			crossOrigin = o.R.Header.Get("Origin")
		}
		o.W.Header().Set("Access-Control-Allow-Origin", crossOrigin)
		o.W.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	o.W.Header().Set("Content-Type", "application/json")
	var ww io.Writer

	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if len(bs) > 1024 {
		o.W.Header().Set("Content-Encoding", "gzip")
		g, _ := gzip.NewWriterLevel(o.W, gzip.BestCompression)
		ww = g
		defer g.Close()
	} else {
		ww = o.W
	}

	o.R.ParseForm()
	callback := o.R.FormValue("callback")
	if callback != "" {
		ww.Write([]byte(callback))
		ww.Write([]byte("("))
		ww.Write(bs)
		ww.Write([]byte(")"))
	} else {
		ww.Write(bs)
	}
	return nil
}

// JSON2 返回json数据，自动识别jsonp
func (o *Context) JSON2(no int, data interface{}) error {
	return o.JSON(&map[string]interface{}{"no": no, "data": data})
}

// HTML 输出HTML代码
func (o *Context) HTML(data string) {
	o.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(o.W, data)
}

// SetTemplateFunc 设置模板函数
func (o *Context) SetTemplateFunc(funcMap template.FuncMap) *Context {
	o.templateFuncMap = funcMap
	return o
}

// LayoutRender 渲染layout模版
func (o *Context) LayoutRender(layout string, data interface{}, tplFiles ...string) {
	o.W.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Delims
	left, right := "{{", "}}"
	if len(o.hst.templateDelims) == 2 {
		left, right = o.hst.templateDelims[0], o.hst.templateDelims[1]
	}

	// layout
	if _, ok := o.hst.layout[layout]; !ok {
		o.HTML("layout not found: " + layout)
		return
	}

	// parse
	tpls := append(o.hst.layout[layout], tplFiles[:]...)
	for k, v := range tpls {
		tpls[k] = o.hst.templatePath + v
	}

	// func
	funcs := template.FuncMap{}
	if o.hst.templateFuncMap != nil {
		for k, v := range o.hst.templateFuncMap {
			funcs[k] = v
		}
	}
	if o.templateFuncMap != nil {
		for k, v := range o.templateFuncMap {
			funcs[k] = v
		}
	}

	// parse
	tpl, err := template.New(layout).Funcs(funcs).Delims(left, right).ParseFiles(tpls[:]...)
	if err != nil {
		o.HTML(err.Error())
		return
	}

	// execute
	if err := tpl.Execute(o.W, data); err != nil {
		o.HTML(err.Error())
		return
	}
}

// RenderFiles 渲染模版
func (o *Context) RenderFiles(delimLeft, delimRight string, data interface{}, tplFiles ...string) {
	o.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.New("").Delims(delimLeft, delimRight).ParseFiles(tplFiles...)
	if err != nil {
		fmt.Fprint(o.W, err)
		return
	}
	name := filepath.Base(tplFiles[len(tplFiles)-1])
	if err := t.ExecuteTemplate(o.W, name, nil); err != nil {
		fmt.Fprint(o.W, err)
	}
}

// RenderContent 渲染内容
func (o *Context) RenderContent(delimLeft, delimRight string, data interface{}, htm ...string) {
	o.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	var err error
	t := template.New("")
	for k, v := range htm {
		t, err = t.New(fmt.Sprintf("%d", k)).Delims(delimLeft, delimRight).Parse(v)
		if err != nil {
			fmt.Fprint(o.W, err)
			return
		}
	}
	if err := t.Delims(delimLeft, delimRight).Execute(o.W, nil); err != nil {
		fmt.Fprint(o.W, err)
	}
}

// SessionSet 设置Session
func (o *Context) SessionSet(key string, value interface{}, expire time.Duration) error {
	return o.session.Set(o, key, value, expire)
}

// SessionGet 读取Session
func (o *Context) SessionGet(key string) (interface{}, error) {
	return o.session.Get(o, key)
}

// SessionDestory 销毁Session
func (o *Context) SessionDestory() error {
	return o.session.Destory(o)
}
