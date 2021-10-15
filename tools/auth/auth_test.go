package auth

import (
	"testing"
)

func TestHashPasswd(t *testing.T) {
	hashVal := HashPasswd("admin", REALM, "hello")
	if hashVal != "7628d9fbecd3683d02276b6176b0ee13" {
		t.Fatalf("hash failed, expect: 7628d9fbecd3683d02276b6176b0ee13, but: %s\n", hashVal)
	}
}
