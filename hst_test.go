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
	h := NewHST(nil)
	h.Favicon()
	h.Static("/abc/", "./")
	h.HandleFunc("/", BasicAuth("u", "p"),
		func(c *Context) {
			c.JSON(msg, false)
		}, func(c *Context) {
			fmt.Fprint(c.W, msg)
		})
	h.HandlePfx("/ssl.pfx", path+domain+".ssl.pfx")
	go h.ListenHTTP(":8080")

	time.Sleep(time.Millisecond * 100)

	{
		res, err := HTTPGet("http://u:p@127.0.0.1:8080")
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != `"`+msg+`"` {
			t.Fatal(string(res))
		}
	}
	{
		res, err := HTTPGet("http://127.0.0.1:8080/abc/LICENSE")
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 1060 {
			t.Fatal(string(res))
		}
	}
	{
		res, err := HTTPGet("http://127.0.0.1:8080/favicon.ico")
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 198 {
			t.Fatal(string(res))
		}
	}
	{
		res, err := HTTPGet("http://127.0.0.1:8080/ssl.pfx")
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
	go h.ListenHTTPS(":8081", path+domain+".ssl.crt", path+domain+".ssl.key")

	time.Sleep(time.Millisecond * 100)

	{
		res, err := HTTPSGet("https://127.0.0.1:8081")
		if err != nil {
			t.Fatal(err)
		}
		if string(res) == msg {
			t.Fatal(string(res))
		}
	}

	{
		res, err := HTTPSGet("https://u:p@127.0.0.1:8081")
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != msg {
			t.Fatal(string(res))
		}
	}
}

func TestNewTLSServer(t *testing.T) {
	h := NewHST(nil)
	h.HandleFunc("/",
		func(c *Context) {
			fmt.Fprint(c.W, msg)
			c.Close()
		}, func(c *Context) {
			fmt.Fprint(c.W, msg)
		})
	go h.ListenTLS(":8082", path+domain+".ca.crt", path+domain+".ssl.crt", path+domain+".ssl.key")

	time.Sleep(time.Millisecond * 100)

	res, err := TLSSGet("https://127.0.0.1:8082", path+domain+".ca.crt", path+domain+".ssl.crt", path+domain+".ssl.key")
	if err != nil {
		t.Fatal(err)
	}
	if string(res) != msg {
		t.Fatal(string(res))
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
