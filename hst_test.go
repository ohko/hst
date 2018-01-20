package hst

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
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
	h, _ := NewHTTPServer(":8080")
	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, msg)
	})
	h.HandleFunc("/ssl.pfx", HandlePfx(path+domain+".ssl.pfx"))
	go h.Listen()

	{
		res, err := HTTPGet("http://127.0.0.1:8080")
		if err != nil {
			t.Fatal(err)
		}
		if string(res) != msg {
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
}

func TestNewHTTPSServer(t *testing.T) {
	h, _ := NewHTTPSServer(":8081", path+domain+".ssl.crt", path+domain+".ssl.key")
	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, msg)
	})
	go h.Listen()

	res, err := HTTPSGet("https://127.0.0.1:8081")
	if err != nil {
		t.Fatal(err)
	}
	if string(res) != msg {
		t.Fatal(string(res))
	}
}

func TestNewTLSServer(t *testing.T) {
	h, err := NewTLSServer(":8082", path+domain+".ca.crt", path+domain+".ssl.crt", path+domain+".ssl.key")
	if err != nil {
		t.Fatal(err)
	}
	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, msg)
	})
	go h.Listen()

	res, err := TLSSGet("https://127.0.0.1:8082", path+domain+".ca.crt", path+domain+".ssl.crt", path+domain+".ssl.key")
	if err != nil {
		t.Fatal(err)
	}
	if string(res) != msg {
		t.Fatal(string(res))
	}
}
