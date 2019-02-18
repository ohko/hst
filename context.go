package hst

import (
	"compress/gzip"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
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
	templateDelims  []string
	templateFuncMap template.FuncMap
}

// Close 结束后面的流程
func (o *Context) Close() {
	o.close = true
	panic(&hstError{"end"})
}

// JSON 返回json数据，自动识别jsonp
func (o *Context) JSON(statusCode int, data interface{}) error {
	defer o.Close()
	o.W.WriteHeader(statusCode)

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
func (o *Context) JSON2(statusCode int, no int, data interface{}) error {
	return o.JSON(statusCode, &map[string]interface{}{"no": no, "data": data})
}

// HTML 输出HTML代码
func (o *Context) HTML(statusCode int, name string, data interface{}) {
	defer o.Close()
	o.W.WriteHeader(statusCode)
	o.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	o.hst.template.ExecuteTemplate(o.W, name, data)
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
