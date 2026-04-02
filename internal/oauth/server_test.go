package oauth_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/zhenqiang/mailbox-cli/internal/oauth"
)

func TestCallbackCapture(t *testing.T) {
	srv, err := oauth.NewCallbackServer()
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(50 * time.Millisecond)
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/callback?code=testcode", srv.Port()))
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
	code, err := srv.WaitForCode(5 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if code != "testcode" {
		t.Fatalf("got %q", code)
	}
}

func TestCallbackTimeout(t *testing.T) {
	srv, err := oauth.NewCallbackServer()
	if err != nil {
		t.Fatal(err)
	}
	_, err = srv.WaitForCode(100 * time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
