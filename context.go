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
	hst   *HST
	W     http.ResponseWriter
	R     *http.Request
	close bool
}

// Close 结束后面的流程
func (o *Context) Close() {
	o.close = true
}

// JSON 返回json数据，自动识别jsonp
func (o *Context) JSON(data interface{}) error {
	defer o.Close()
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
func (o *Context) SessionSet(key string, value interface{}, expire time.Duration) {
	o.hst.sessionLock.Lock()
	defer o.hst.sessionLock.Unlock()

	ck, err := o.R.Cookie(SESSIONKEY)
	if err != nil {
		ck = &http.Cookie{
			Name:     SESSIONKEY,
			Value:    MakeGUID(),
			HttpOnly: true,
		}
		o.R.Header.Set("Cookie", ck.String())
		http.SetCookie(o.W, ck)
	}

	if v, ok := o.hst.sessionData[ck.Value]; ok {
		if vv, ok := (*v)[key]; ok {
			vv.data = value
			vv.expire = time.Now().Add(expire)
			return
		}
		(*v)[key] = &sessionData{data: value, expire: time.Now().Add(expire)}
		return
	}

	data := &sessionData{data: value, expire: time.Now().Add(expire)}
	sess := &map[string]*sessionData{key: data}
	o.hst.sessionData[ck.Value] = sess
}

// SessionGet 读取Session
func (o *Context) SessionGet(key string) interface{} {
	ck, err := o.R.Cookie(SESSIONKEY)
	if err != nil {
		return nil
	}

	o.hst.sessionLock.RLock()
	defer o.hst.sessionLock.RUnlock()

	if v, ok := o.hst.sessionData[ck.Value]; ok {
		if vv, ok := (*v)[key]; ok {
			if vv.expire.Sub(time.Now()) > 0 {
				return vv.data
			}
		}
	}

	return nil
}

// SessionDestory 销毁Session
func (o *Context) SessionDestory() interface{} {
	ck, err := o.R.Cookie(SESSIONKEY)
	if err != nil {
		return nil
	}

	o.hst.sessionLock.Lock()
	defer o.hst.sessionLock.Unlock()

	if v, ok := o.hst.sessionData[ck.Value]; ok {
		for kk := range *v {
			delete(*v, kk)
		}
		delete(o.hst.sessionData, ck.Value)
	}
	ck.Expires = time.Now().Add(-1)
	http.SetCookie(o.W, ck)

	return nil
}

func (o *Context) cleanSession() {
	for {
		time.Sleep(time.Minute)
		o.hst.sessionLock.Lock()
		for k, v := range o.hst.sessionData {
			for kk, vv := range *v {
				if vv.expire.Sub(time.Now()) <= 0 {
					delete(*v, kk)
				}
			}
			if len(*v) == 0 {
				delete(o.hst.sessionData, k)
			}
		}
		o.hst.sessionLock.Unlock()
	}
}
