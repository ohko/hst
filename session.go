package hst

import (
	"net/http"
	"sync"
	"time"
)

// const
const SESSIONKEY = "HST_SESSION"

// Session ...
type Session interface {
	Set(c *Context, key string, value interface{}, expire time.Duration)
	Get(c *Context, key string) interface{}
	Destory(c *Context)
}

// MemorySession ...
type MemorySession struct {
	lock sync.RWMutex
	data map[string]*map[string]*memSessionData
}
type memSessionData struct {
	data   interface{}
	expire time.Time
}

// NewMemorySession ...
func NewMemorySession() Session {
	o := new(MemorySession)
	o.data = make(map[string]*map[string]*memSessionData)
	go o.cleanSession()
	return o
}

// Set 设置Session
func (o *MemorySession) Set(c *Context, key string, value interface{}, expire time.Duration) {
	o.lock.Lock()
	defer o.lock.Unlock()

	ck, err := c.R.Cookie(SESSIONKEY)
	if err != nil {
		ck = &http.Cookie{
			Name:     SESSIONKEY,
			Value:    MakeGUID(),
			HttpOnly: true,
		}
		c.R.Header.Set("Cookie", ck.String())
		http.SetCookie(c.W, ck)
	}

	if v, ok := o.data[ck.Value]; ok {
		if vv, ok := (*v)[key]; ok {
			vv.data = value
			vv.expire = time.Now().Add(expire)
			return
		}
		(*v)[key] = &memSessionData{data: value, expire: time.Now().Add(expire)}
		return
	}

	data := &memSessionData{data: value, expire: time.Now().Add(expire)}
	sess := &map[string]*memSessionData{key: data}
	o.data[ck.Value] = sess
}

// Get 读取Session
func (o *MemorySession) Get(c *Context, key string) interface{} {
	ck, err := c.R.Cookie(SESSIONKEY)
	if err != nil {
		return nil
	}

	o.lock.RLock()
	defer o.lock.RUnlock()

	if v, ok := o.data[ck.Value]; ok {
		if vv, ok := (*v)[key]; ok {
			if vv.expire.Sub(time.Now()) > 0 {
				return vv.data
			}
		}
	}

	return nil
}

// Destory 销毁Session
func (o *MemorySession) Destory(c *Context) {
	ck, err := c.R.Cookie(SESSIONKEY)
	if err != nil {
		return
	}

	o.lock.Lock()
	defer o.lock.Unlock()

	if v, ok := o.data[ck.Value]; ok {
		for kk := range *v {
			delete(*v, kk)
		}
		delete(o.data, ck.Value)
	}
	ck.Expires = time.Now().Add(-1)
	http.SetCookie(c.W, ck)
}

func (o *MemorySession) cleanSession() {
	for {
		time.Sleep(time.Minute)
		o.lock.Lock()
		for k, v := range o.data {
			for kk, vv := range *v {
				if vv.expire.Sub(time.Now()) <= 0 {
					delete(*v, kk)
				}
			}
			if len(*v) == 0 {
				delete(o.data, k)
			}
		}
		o.lock.Unlock()
	}
}
