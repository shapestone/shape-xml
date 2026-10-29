package xml

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// xxePayloads returns documents that try to make a parser fetch url and path,
// through general, parameter and nested entities, and the expected verdict.
func xxePayloads(url, path string) []struct {
	name, doc string
	valid     bool
} {
	return []struct {
		name, doc string
		valid     bool
	}{
		{"file entity used", `<?xml version="1.0"?><!DOCTYPE d [<!ENTITY x SYSTEM "file://` + path + `">]><d>&x;</d>`, false},
		{"file entity declared only", `<!DOCTYPE d [<!ENTITY x SYSTEM "file://` + path + `">]><d>ok</d>`, true},
		{"http entity used", `<!DOCTYPE d [<!ENTITY x SYSTEM "` + url + `/e">]><d>&x;</d>`, false},
		{"http entity declared only", `<!DOCTYPE d [<!ENTITY x SYSTEM "` + url + `/e">]><d>ok</d>`, true},
		{"parameter entity", `<!DOCTYPE d [<!ENTITY % p SYSTEM "` + url + `/p"> %p;]><d>ok</d>`, true},
		{"parameter entity in attribute default", `<!DOCTYPE d [<!ENTITY % p SYSTEM "` + url + `/p"><!ATTLIST d a CDATA "%p;">]><d/>`, true},
		{"external DTD", `<!DOCTYPE d SYSTEM "` + url + `/d.dtd"><d/>`, true},
		{"external DTD by public id", `<!DOCTYPE d PUBLIC "-//X//Y" "` + url + `/d.dtd"><d/>`, true},
		{"external DTD with file", `<!DOCTYPE d SYSTEM "file://` + path + `"><d/>`, true},
		{"entity in attribute", `<!DOCTYPE d [<!ENTITY x SYSTEM "` + url + `/a">]><d a="&x;"/>`, false},
		{"internal entity used", `<!DOCTYPE d [<!ENTITY x "plain">]><d>&x;</d>`, false},
		{"billion laughs declared", laughs(false), true},
		{"billion laughs used", laughs(true), false},
	}
}

func laughs(use bool) string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE lolz [<!ENTITY lol "lol">`)
	prev := "lol"
	for i := 1; i <= 9; i++ {
		name := "lol" + string(rune('0'+i))
		sb.WriteString(`<!ENTITY ` + name + ` "` + strings.Repeat("&"+prev+";", 10) + `">`)
		prev = name
	}
	sb.WriteString(`]><lolz>`)
	if use {
		sb.WriteString("&" + prev + ";")
	}
	sb.WriteString(`</lolz>`)
	return sb.String()
}

// TestValidate_XXEPayloads proves that XXE-style payloads get the documented
// verdict and that no HTTP request reaches a local server.
func TestValidate_XXEPayloads(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = w.Write([]byte("<!ENTITY y 'z'>"))
	}))
	defer srv.Close()

	secret := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(secret, []byte("TOP-SECRET"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, tc := range xxePayloads(srv.URL, secret) {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			err := Validate(tc.doc)
			if (err == nil) != tc.valid {
				t.Errorf("Validate = %v, want valid=%v", err, tc.valid)
			}
			if err != nil && strings.Contains(err.Error(), "TOP-SECRET") {
				t.Errorf("error leaks file contents: %v", err)
			}
			if err := ValidateReader(strings.NewReader(tc.doc)); (err == nil) != tc.valid {
				t.Errorf("ValidateReader = %v, want valid=%v", err, tc.valid)
			}
			if d := time.Since(start); d > time.Second {
				t.Errorf("took %v; entities must not be expanded", d)
			}
		})
	}
	if n := atomic.LoadInt32(&hits); n != 0 {
		t.Errorf("local server received %d request(s); Validate must not use the network", n)
	}
}

// A dialer-level check: nothing is listening that Validate could reach, and
// the payload names a closed port. If Validate tried to connect it would not
// matter for the verdict, so also assert the verdict and that no connection is
// made to a listener we control.
func TestValidate_XXENoConnection(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("cannot listen:", err)
	}
	defer ln.Close()
	accepted := make(chan struct{}, 1)
	go func() {
		if c, err := ln.Accept(); err == nil {
			accepted <- struct{}{}
			c.Close()
		}
	}()

	doc := `<!DOCTYPE d [<!ENTITY % p SYSTEM "http://` + ln.Addr().String() + `/x"> %p;]><d/>`
	if err := Validate(doc); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	select {
	case <-accepted:
		t.Fatal("Validate opened a connection")
	case <-time.After(200 * time.Millisecond):
	}
}
