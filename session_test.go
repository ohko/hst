package hst

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewSessionMemory(t *testing.T) {
	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	w := mywrite{}
	c := &Context{W: &responseWriterWithLength{w, 0}, R: r}

	s := NewSessionMemory("HST_SESSION")
	if v, err := s.Get(c, "a"); err == nil || v != nil {
		t.Error(v, err)
	}
	s.Set(c, "", "/", "a", "A", time.Second)
	if v, err := s.Get(c, "a"); err != nil || v.(string) != "A" {
		t.Error(v, err)
	}
	time.Sleep(time.Second)
	if v, err := s.Get(c, "a"); err == nil || v != nil {
		t.Error(v, err)
	}
}

func TestNewSessionFile(t *testing.T) {
	r := httptest.NewRequest("GET", "/", strings.NewReader(""))
	w := mywrite{}
	c := &Context{W: &responseWriterWithLength{w, 0}, R: r}

	s := NewSessionFile("HST_SESSION", os.TempDir()+"HST", time.Minute)
	if v, err := s.Get(c, "a"); err == nil || v != nil {
		t.Error(v, err)
	}
	s.Set(c, "", "/", "a", "A", time.Second)
	if v, err := s.Get(c, "a"); err != nil || v.(string) != "A" {
		t.Error(v, err)
	}
	time.Sleep(time.Second)
	if v, err := s.Get(c, "a"); err == nil || v != nil {
		t.Error(v, err)
	}
}

type mywrite struct{}

func (o mywrite) Header() http.Header        { return http.Header{} }
func (o mywrite) Write([]byte) (int, error)  { return 0, nil }
func (o mywrite) WriteHeader(statusCode int) {}
