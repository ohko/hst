package hst

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Context 上下文数据
type Context struct {
	W     http.ResponseWriter
	R     *http.Request
	close bool
}

// Close 结束后面的流程
func (o *Context) Close() {
	o.close = true
}

// JSON 返回json数据，自动识别jsonp
func (o *Context) JSON(data interface{}, gz bool) error {
	defer o.Close()

	o.W.Header().Set("Content-Type", "application/json")
	var ww io.Writer
	if gz {
		o.W.Header().Set("Content-Encoding", "gzip")
		g := gzip.NewWriter(o.W)
		ww = g
		defer g.Close()
	} else {
		ww = o.W
	}

	// js, err := jsoniter.MarshalToString(data)
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}
	js := string(bs)

	o.R.ParseForm()
	callback := o.R.FormValue("callback")
	if callback != "" {
		fmt.Fprint(ww, callback+"(", js, ")")
	} else {
		fmt.Fprint(ww, js)
	}
	return nil
}

// JSON2 返回json数据，自动识别jsonp
func (o *Context) JSON2(no int, data interface{}, gz bool) error {
	return o.JSON(&map[string]interface{}{"no": no, "data": data}, gz)
}
