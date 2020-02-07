package main

import (
	"testing"
)

func TestMakeKey(t *testing.T) {
	passwd := "p4ssw0rdE"
	email := "nobody@example.com"
	iterations := 5000

	want := ""
	got, _ := makeKey(passwd, email, iterations)

	if got != want {
		t.Errorf("got %v\nwant %v\n", got, want)
	}
}
