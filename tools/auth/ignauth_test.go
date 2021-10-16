package auth

import (
	"testing"
)

func TestParseIgnoreAuth(t *testing.T) {
	ParseIgnoreAuth(nil)

	in := []byte(`
/*.html
/js
/css
`)
	ignAuth := ParseIgnoreAuth(in)
	if !ignAuth.Match("/test.html") {
		t.Fatal("expect true")
	}
	if ignAuth.Match("/html/*.html") {
		t.Fatal("expect false")
	}
	if !ignAuth.Match("/js") {
		t.Fatal("expect true")
	}
	if !ignAuth.Match("/js/test.js") {
		t.Fatal("expect true")
	}
	if ignAuth.Match("/doc/js") {
		t.Fatal("expect false")
	}
}
