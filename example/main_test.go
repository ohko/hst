package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"testing"
)

func Test_main(t *testing.T) {
	fs, err := ioutil.ReadDir("./")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range fs {
		fmt.Println(v.IsDir(), v.Name())
	}

	tpl, err := template.ParseGlob("./template/**/*")
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range tpl.Templates() {
		fmt.Println(v.Name())
	}
}
