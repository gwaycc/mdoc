package auth

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gwaylib/errors"
)

func TestUserInfo(t *testing.T) {
	InitDB("./user_test/mdoc.db")
	defer os.RemoveAll("./user_test")

	username := fmt.Sprintf("%d", time.Now().UnixNano())
	input := &UserInfo{ID: username, Passwd: "testing", NickName: "testing"}
	if _, err := GetUser(username); !errors.ErrNoData.Equal(err) {
		t.Fatal("need data not exist, but: ", err)
	}
	if err := AddUser(input); err != nil {
		t.Fatal(err)
	}
	// checksum the result
	output, err := GetUser(username)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%+v", input) != fmt.Sprintf("%+v", output) {
		t.Fatalf("expect %+v, but: %+v\n", input, output)
	}
}
