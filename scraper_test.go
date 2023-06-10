package main

import (
	"net/http"
	"testing"
)

func TestCheckResponse(t *testing.T) {
	url := "https://www.google.com/"
	resp, _ := http.Get(url)
	got := CheckRespStatus(resp)
	if got != nil {
		t.Errorf("got error, want no error")
	}
}
