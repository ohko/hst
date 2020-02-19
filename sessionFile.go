package hst

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// SessionFile ...
type SessionFile struct {
	cookieName   string
	cookiePath   string
	cookieDomain string
	cookieExpire time.Duration
	path         string
	lock         sync.RWMutex
	update       map[string]time.Time // 更新的session
	maxExpire    time.Duration        // 文件过期时间
}

// NewSessionFile ...
func NewSessionFile(cookieDomain, cookiePath, cookieName, path string, maxExpire time.Duration) Session {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	os.MkdirAll(path, 0755)
	o := new(SessionFile)
	o.cookieName = cookieName
	o.cookieDomain = cookieDomain
	o.cookiePath = cookiePath
	o.path = path
	o.maxExpire = maxExpire
	o.cookieExpire = maxExpire
	go o.cleanSession()
	return o
}

// GetCookieExpire 获取cookie的过期时间
func (o *SessionFile) GetCookieExpire() time.Duration {
	return o.cookieExpire
}

// Set 设置Session
func (o *SessionFile) Set(c *Context, key string, value interface{}, expire time.Duration) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	ck, err := c.R.Cookie(o.cookieName)
	if err != nil {
		ck = &http.Cookie{
			Domain:   o.cookieDomain,
			Path:     o.cookiePath,
			Name:     o.cookieName,
			Value:    MakeGUID(),
			MaxAge:   int(o.maxExpire.Seconds()),
			Expires:  time.Now().Add(o.cookieExpire),
			HttpOnly: true,
		}
		c.R.AddCookie(ck)
		http.SetCookie(c.W, ck)
	}

	var data map[string]sessionData
	{ // 读取
		if bs, err := ioutil.ReadFile(o.path + ck.Value); err == nil {
			if err := json.Unmarshal(bs, &data); err != nil {
				data = make(map[string]sessionData)
			}
		} else {
			data = make(map[string]sessionData)
		}
	}

	data[key] = sessionData{Data: value, Expire: time.Now().Add(expire)}

	{ // 保存
		bs, err := json.Marshal(&data)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(o.path+ck.Value, bs, 0644); err != nil {
			return err
		}
	}

	return nil
}

// Get 读取Session
func (o *SessionFile) Get(c *Context, key string) (interface{}, error) {
	ck, err := c.R.Cookie(o.cookieName)
	if err != nil {
		return nil, err
	}

	o.lock.RLock()
	defer o.lock.RUnlock()

	var data map[string]sessionData
	// 读取
	bs, err := ioutil.ReadFile(o.path + ck.Value)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bs, &data); err != nil {
		return nil, err
	}

	defer func() {
		os.Chtimes(o.path+ck.Value, time.Now(), time.Now())
	}()

	if v, ok := data[key]; ok {
		if v.Expire.Sub(time.Now()) > 0 {
			return v.Data, nil
		}
		return nil, errors.New("expire")
	}

	return nil, errors.New("not found")
}

// Destory 销毁Session
func (o *SessionFile) Destory(c *Context) error {
	ck, err := c.R.Cookie(o.cookieName)
	if err != nil {
		return err
	}

	ck.Expires = time.Now().Add(-1)
	http.SetCookie(c.W, ck)

	o.lock.Lock()
	defer o.lock.Unlock()

	if err := os.Remove(o.path + ck.Value); err != nil {
		return err
	}

	return nil
}

func (o *SessionFile) cleanSession() {
	for {
		time.Sleep(time.Minute)
		func() {
			o.lock.Lock()
			defer o.lock.Unlock()

			fi, err := ioutil.ReadDir(o.path)
			if err != nil {
				return
			}
			for _, v := range fi {
				if time.Now().Sub(v.ModTime()) > o.maxExpire {
					os.Remove(o.path + v.Name())
				}
			}
		}()
	}
}
