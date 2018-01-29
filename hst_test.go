package hst

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

var pass1 = "123"
var pass2 = "123"
var pass3 = "123"
var path = "/tmp/hst-ssl/"
var domain = "test"
var email = "hk@cdeyun.com"
var msg = "Hello world!"

func TestMakeTLSFile(t *testing.T) {
	if !MakeTLSFile(pass1, pass2, pass3, path, domain, email) {
		t.Fatal("make tls error!")
	}
}

func TestNewHTTPServer(t *testing.T) {
	hs := &Handlers{
		"/": []HandlerFunc{
			func(c *Context) {
				c.JSON(msg)
			}, func(c *Context) {
				fmt.Fprint(c.W, msg)
			},
		},
	}

	h := NewHST(hs)
	h.Favicon()
	h.Static("/abc/", "./")
	h.HandlePfx("/ssl.pfx", path+domain+".ssl.pfx")
	go h.ListenHTTP(":8280")

	time.Sleep(time.Millisecond * 100)

	{
		res, _, err := HTTPGet("http://u:p@127.0.0.1:8280", "")
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != `"`+msg+`"` {
			t.Fatal(string(res))
		}
	}
	{
		res, _, err := HTTPGet("http://127.0.0.1:8280/abc/LICENSE", "")
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 1060 {
			t.Fatal(string(res))
		}
	}
	{
		res, _, err := HTTPGet("http://127.0.0.1:8280/favicon.ico", "")
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 198 {
			t.Fatal(string(res))
		}
	}
	{
		res, _, err := HTTPGet("http://127.0.0.1:8280/ssl.pfx", "")
		if err != nil {
			t.Fatal(err)
		}
		bs, err := ioutil.ReadFile(path + domain + ".ssl.pfx")
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != string(bs) {
			t.Fatal(string(res))
		}
	}

	log.Println("wait ctrl+c ...")
	// Shutdown([]*HST{h}, time.Second*5)
}

func TestNewHTTPSServer(t *testing.T) {
	h := NewHST(nil)
	h.HandleFunc("/", BasicAuth("u", "p"), func(c *Context) {
		fmt.Fprint(c.W, msg)
	})
	go h.ListenHTTPS(":8281", path+domain+".ssl.crt", path+domain+".ssl.key")

	time.Sleep(time.Millisecond * 100)

	{
		res, _, err := HTTPSGet("https://127.0.0.1:8281", "")
		if err != nil {
			t.Fatal(err)
		}
		if string(res) == msg {
			t.Fatal(string(res))
		}
	}

	{
		res, _, err := HTTPSGet("https://u:p@127.0.0.1:8281", "")
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != msg {
			t.Fatal(string(res))
		}
	}
}

func TestNewTLSServer(t *testing.T) {
	httpAndTLS := NewHandlers()
	httpAndTLS.HandlerFunc("/hANDt", func(c *Context) {
		fmt.Fprint(c.W, msg)
	})
	h := NewHST(&httpAndTLS)
	h.HandleFunc("/",
		func(c *Context) {
			fmt.Fprint(c.W, msg)
			c.Close()
		}, func(c *Context) {
			fmt.Fprint(c.W, msg)
		})
	h.HandleFunc("/SetSession", func(c *Context) {
		c.SessionSet("a", msg, time.Minute)
	})
	h.HandleFunc("/GetSession", func(c *Context) {
		v := c.SessionGet("a")
		if v == nil {
			fmt.Fprint(c.W, "...")
			return
		}
		fmt.Fprint(c.W, v.(string))
	})
	go h.ListenTLS(":8282", path+domain+".ca.crt", path+domain+".ssl.crt", path+domain+".ssl.key")

	h2 := NewHST(&httpAndTLS)
	go h2.ListenHTTP(":8283")

	time.Sleep(time.Millisecond * 200)

	{
		res, _, err := HTTPGet("http://127.0.0.1:8283/hANDt", "")
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != msg {
			t.Fatal(string(res))
		}
	}
	{
		res, _, err := TLSSGet("https://127.0.0.1:8282/hANDt", path+domain+".ca.crt", path+domain+".ssl.crt", path+domain+".ssl.key", "")
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != msg {
			t.Fatal(string(res))
		}
	}
	{
		res, _, err := TLSSGet("https://127.0.0.1:8282", path+domain+".ca.crt", path+domain+".ssl.crt", path+domain+".ssl.key", "")
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != msg {
			t.Fatal(string(res))
		}
	}
	{
		_, cs, _ := TLSSGet("https://127.0.0.1:8282/SetSession", path+domain+".ca.crt", path+domain+".ssl.crt", path+domain+".ssl.key", "")
		cookie := ""
		for _, v := range cs {
			if v.Name == SESSIONKEY {
				cookie = v.Value
				break
			}
		}
		if cookie != "" {
			log.Println(cookie)
		}
		res, _, err := TLSSGet("https://127.0.0.1:8282/GetSession", path+domain+".ca.crt", path+domain+".ssl.crt", path+domain+".ssl.key", SESSIONKEY+"="+cookie)
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != msg {
			t.Fatal(string(res))
		}
	}
}

func BenchmarkA(t *testing.B) {
	t.ResetTimer()
	t.SetBytes(100)
	for i := 0; i < t.N; i++ {
		panicRecover()
	}
}

func BenchmarkB(t *testing.B) {
	t.ResetTimer()
	t.SetBytes(100)
	for i := 0; i < t.N; i++ {
		noPanicRecover()
	}
}

func panicRecover() {
	defer func() {
		recover()
	}()
	panic(errors.New("nil"))
}

func noPanicRecover() error {
	defer func() {
		recover()
	}()
	if err := do(); err != nil {
		return err
	}
	return nil
}

func do() error {
	return errors.New("nil")
}
