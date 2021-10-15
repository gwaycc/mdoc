package repo

import (
	"github.com/mitchellh/go-homedir"
)

func ExpandPath(repoDir string) string {
	r, err := homedir.Expand(repoDir)
	if err != nil {
		panic(err)
	}
	return r
}
