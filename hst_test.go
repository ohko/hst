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

func Example_main() {
	s := New(nil)
	s.ListenHTTP(":8080")
}

func TestMakeTLSFile(t *testing.T) {
	if !MakeTLSFile(pass1, pass2, pass3, path, domain, email) {
		t.Fatal("make tls error!")
	}
}

func TestNewHTTPServer(t *testing.T) {
	hs := &Handlers{
		"/": []HandlerFunc{
			func(c *Context) {
				c.JSON(200, msg)
			}, func(c *Context) {
				fmt.Fprint(c.W, msg)
			},
		},
	}

	h := New(hs)
	h.Favicon()
	h.Static("/abc/", "./")
	h.HandlePfx("/ssl.pfx", path+domain+".ssl.pfx")
	go h.ListenHTTP(":8280")

	time.Sleep(time.Millisecond * 100)

	{
		res, _, err := Request("GET", "http://u:p@127.0.0.1:8280", "", "", nil)
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != `"`+msg+`"` {
			t.Fatal(string(res))
		}
	}
	{
		res, _, err := Request("GET", "http://127.0.0.1:8280/abc/LICENSE", "", "", nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 1060 {
			t.Fatal(string(res))
		}
	}
	{
		res, _, err := Request("GET", "http://127.0.0.1:8280/favicon.ico", "", "", nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 198 {
			t.Fatal(string(res))
		}
	}
	{
		res, _, err := Request("GET", "http://127.0.0.1:8280/ssl.pfx", "", "", nil)
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
	h := New(nil)
	h.HandleFunc("/", BasicAuth("u", "p"), func(c *Context) {
		fmt.Fprint(c.W, msg)
	})
	go h.ListenHTTPS(":8281", path+domain+".ssl.crt", path+domain+".ssl.key")

	time.Sleep(time.Millisecond * 100)

	{
		res, _, err := Request("GET", "https://127.0.0.1:8281", "", "", nil)
		if err != nil {
			t.Fatal(err)
		}
		if string(res) == msg {
			t.Fatal(string(res))
		}
	}

	{
		res, _, err := Request("GET", "https://u:p@127.0.0.1:8281", "", "", nil)
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
	h := New(&httpAndTLS)
	h.SetSession(NewSessionMemory("", "/", "HST_SESSION", time.Minute))
	// h.SetSession(NewSessionFile("/tmp/hstSession", time.Hour))
	h.HandleFunc("/",
		func(c *Context) {
			fmt.Fprint(c.W, msg)
			c.Close()
		}, func(c *Context) {
			fmt.Fprint(c.W, msg)
		})
	h.HandleFunc("/SetSession", func(c *Context) {
		c.SessionSet("a", msg)
	})
	h.HandleFunc("/GetSession", func(c *Context) {
		v, _ := c.SessionGet("a")
		if v == nil {
			fmt.Fprint(c.W, "...")
			return
		}
		fmt.Fprint(c.W, v.(string))
	})
	go h.ListenTLS(":8282", path+domain+".ca.crt", path+domain+".ssl.crt", path+domain+".ssl.key")

	h2 := New(&httpAndTLS)
	go h2.ListenHTTP(":8283")

	time.Sleep(time.Millisecond * 200)

	{
		res, _, err := Request("GET", "http://127.0.0.1:8283/hANDt", "", "", nil)
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != msg {
			t.Fatal(string(res))
		}
	}
	{
		res, _, err := RequestTLS("GET", "https://127.0.0.1:8282/hANDt", path+domain+".ca.crt", path+domain+".ssl.crt", path+domain+".ssl.key", "", "")
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != msg {
			t.Fatal(string(res))
		}
	}
	{
		res, _, err := RequestTLS("GET", "https://127.0.0.1:8282", path+domain+".ca.crt", path+domain+".ssl.crt", path+domain+".ssl.key", "", "")
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != msg {
			t.Fatal(string(res))
		}
	}
	{
		_, cs, _ := RequestTLS("GET", "https://127.0.0.1:8282/SetSession", path+domain+".ca.crt", path+domain+".ssl.crt", path+domain+".ssl.key", "", "")
		cookie := ""
		for _, v := range cs {
			if v.Name == "HST_SESSION" {
				cookie = v.Value
				break
			}
		}
		if cookie != "" {
			log.Println(cookie)
		}
		res, _, err := RequestTLS("GET", "https://127.0.0.1:8282/GetSession", path+domain+".ca.crt", path+domain+".ssl.crt", path+domain+".ssl.key", "HST_SESSION="+cookie, "")
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

func TestShutdown(t *testing.T) {
	s1 := New(nil)
	s2 := New(nil)
	s3 := New(nil)
	go s1.ListenHTTP(":8081")
	go s2.ListenHTTP(":8082")
	go s3.ListenHTTP(":8083")

	Shutdown(time.Second*5, s1, s2, s3)
}
