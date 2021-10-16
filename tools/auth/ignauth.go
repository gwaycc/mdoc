package auth

import (
	"bytes"
	"encoding/csv"
	"path/filepath"
	"strings"
)

type Prefix struct {
	Path   string
	Regexp bool
}

type IgnoreAuth struct {
	prefixes []Prefix
}

func ParseIgnoreAuth(data []byte) *IgnoreAuth {
	r := csv.NewReader(bytes.NewReader(data))
	r.Comment = '#'
	record, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	prefixes := []Prefix{}
	for _, r := range record {
		if len(r) > 0 {
			prefixes = append(prefixes, Prefix{Path: r[0], Regexp: strings.Contains(r[0], "*")})
		}
	}
	return NewIgnoreAuth(prefixes)
}

func NewIgnoreAuth(prefix []Prefix) *IgnoreAuth {
	return &IgnoreAuth{prefixes: prefix}
}

func (n *IgnoreAuth) Match(path string) bool {
	for _, p := range n.prefixes {
		if p.Regexp {
			if ok, _ := filepath.Match(p.Path, path); ok {
				return true
			}
			continue
		}

		if strings.HasPrefix(path, p.Path) {
			return true
		}
	}
	return false
}
