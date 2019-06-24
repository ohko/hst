package hst

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

// SessionMemory ...
type SessionMemory struct {
	cookieName string
	lock       sync.RWMutex
	data       map[string]*map[string]*memSessionData
}

// NewSessionMemory ...
func NewSessionMemory(cookieName string) Session {
	o := new(SessionMemory)
	o.cookieName = cookieName
	o.data = make(map[string]*map[string]*memSessionData)
	go o.cleanSession()
	return o
}

// Set 设置Session
func (o *SessionMemory) Set(c *Context, domain, path, key string, value interface{}, expire time.Duration) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	ck, err := c.R.Cookie(o.cookieName)
	if err != nil {
		ck = &http.Cookie{
			Domain:   domain,
			Path:     path,
			Name:     o.cookieName,
			Value:    MakeGUID(),
			HttpOnly: true,
		}
		c.R.Header.Set("Cookie", ck.String())
		http.SetCookie(c.W, ck)
	}

	if v, ok := o.data[ck.Value]; ok {
		if vv, ok := (*v)[key]; ok {
			vv.Data = value
			vv.Expire = time.Now().Add(expire)
			return nil
		}
		(*v)[key] = &memSessionData{Data: value, Expire: time.Now().Add(expire)}
		return nil
	}

	data := &memSessionData{Data: value, Expire: time.Now().Add(expire)}
	sess := &map[string]*memSessionData{key: data}
	o.data[ck.Value] = sess
	return nil
}

// Get 读取Session
func (o *SessionMemory) Get(c *Context, key string) (interface{}, error) {
	ck, err := c.R.Cookie(o.cookieName)
	if err != nil {
		return nil, err
	}

	o.lock.RLock()
	defer o.lock.RUnlock()

	if v, ok := o.data[ck.Value]; ok {
		if vv, ok := (*v)[key]; ok {
			if vv.Expire.Sub(time.Now()) > 0 {
				return vv.Data, nil
			}
			return nil, errors.New("expire")
		}
	}

	return nil, errors.New("not found")
}

// Destory 销毁Session
func (o *SessionMemory) Destory(c *Context) error {
	ck, err := c.R.Cookie(o.cookieName)
	if err != nil {
		return err
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
	return nil
}

func (o *SessionMemory) cleanSession() {
	for {
		time.Sleep(time.Minute)
		func() {
			o.lock.Lock()
			defer o.lock.Unlock()
			for k, v := range o.data {
				for kk, vv := range *v {
					if vv.Expire.Sub(time.Now()) <= 0 {
						delete(*v, kk)
					}
				}
				if len(*v) == 0 {
					delete(o.data, k)
				}
			}
		}()
	}
}
