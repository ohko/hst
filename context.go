package hst

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"
)

// Context 上下文数据
type Context struct {
	hst    *HST
	W      *responseWriterWithLength
	R      *http.Request
	status int
	close  bool

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
	o.status = statusCode

	if o.hst.CrossOrigin != "" {
		crossOrigin := o.hst.CrossOrigin
		if o.hst.CrossOrigin == "*" {
			crossOrigin = o.R.Header.Get("Origin")
		}
		o.W.Header().Set("Access-Control-Allow-Origin", crossOrigin)
		// o.W.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	o.W.Header().Set("Content-Type", "application/json")

	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// var ww io.Writer
	// Accept-Encoding: deflate, gzip
	// if len(bs) > 1024 && strings.Contains(o.R.Header.Get("Accept-Encoding"), "gzip") {
	// 	o.W.Header().Set("Content-Encoding", "gzip")
	// 	g, _ := gzip.NewWriterLevel(o.W, gzip.BestCompression)
	// 	ww = g
	// 	defer g.Close()
	// } else {
	// ww = o.W
	// }

	o.W.WriteHeader(statusCode)

	o.R.ParseForm()
	callback := o.R.FormValue("callback")
	if callback != "" {
		o.W.Write([]byte(callback))
		o.W.Write([]byte("("))
		o.W.Write(bs)
		o.W.Write([]byte(")"))
	} else {
		o.W.Write(bs)
	}
	return nil
}

// JSON2 返回json数据，自动识别jsonp
func (o *Context) JSON2(statusCode int, no int, data interface{}) error {
	return o.JSON(statusCode, &map[string]interface{}{"no": no, "data": data})
}

// HTML 从模版缓存输出HTML模版，需要hst.ParseGlob或hst.ParseFiles
func (o *Context) HTML(statusCode int, name string, data interface{}, names ...string) {
	defer o.Close()
	o.status = statusCode
	o.W.WriteHeader(statusCode)
	o.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	o.hst.template.ExecuteTemplate(o.W, name, data)
}

// HTML2 实时读取模版输出HTML模版，需要hst.SetTemplatePath
// name: 主模版
// names: 需要的其它模版组件
func (o *Context) HTML2(statusCode int, name string, data interface{}, names ...string) {
	defer o.Close()
	o.status = statusCode

	names = append(names, name)
	for k, v := range names {
		names[k] = o.hst.templatePath + v
	}
	tpl, err := template.New(name).
		Delims(o.hst.templateDelims.left, o.hst.templateDelims.right).
		Funcs(o.hst.templateFuncMap).
		ParseFiles(names...)
	if err != nil {
		o.Data(statusCode, err)
	}

	o.W.WriteHeader(statusCode)
	o.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.Execute(o.W, data); err != nil {
		o.Data(statusCode, err)
	}
}

// Data 输出对象数据
func (o *Context) Data(statusCode int, data interface{}) {
	defer o.Close()
	if o.hst.CrossOrigin != "" {
		crossOrigin := o.hst.CrossOrigin
		if o.hst.CrossOrigin == "*" {
			crossOrigin = o.R.Header.Get("Origin")
		}
		o.W.Header().Set("Access-Control-Allow-Origin", crossOrigin)
	}
	o.status = statusCode
	o.W.WriteHeader(statusCode)
	fmt.Fprint(o.W, data)
}

// SessionSet 设置Session，默认30分钟后过期
func (o *Context) SessionSet(key string, value interface{}) error {
	return o.hst.session.Set(o, key, value, time.Minute*30)
}

// SessionSetExpire 设置Session，附带过期时间
func (o *Context) SessionSetExpire(key string, value interface{}, expire time.Duration) error {
	return o.hst.session.Set(o, key, value, expire)
}

// SessionGet 读取Session
func (o *Context) SessionGet(key string) (interface{}, error) {
	return o.hst.session.Get(o, key)
}

// SessionDestory 销毁Session
func (o *Context) SessionDestory() error {
	return o.hst.session.Destory(o)
}

// SetCookie 设置cookie
func (o *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(o.W, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

// Cookie 获取cookie
func (o *Context) Cookie(name string) (string, error) {
	cook, err := o.R.Cookie(name)
	if err != nil {
		return "", err
	}
	return url.QueryUnescape(cook.Value)
}
