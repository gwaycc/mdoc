package auth

import (
	"github.com/gwaylib/database"
	"github.com/gwaylib/errors"
)

type UserInfo struct {
	ID       string `db:"id"`
	Passwd   string `db:"passwd"`
	NickName string `db:"nick_name"`
}

func AddUser(uInfo *UserInfo) error {
	db := GetDB()
	if _, err := database.InsertStruct(db, uInfo, "user_info"); err != nil {
		return errors.As(err)
	}
	return nil
}

func GetUser(username string) (*UserInfo, error) {
	uInfo := &UserInfo{}
	db := GetDB()
	if err := database.QueryStruct(db, uInfo, `SELECT * FROM user_info WHERE id=?`, username); err != nil {
		return nil, errors.As(err, username)
	}
	return uInfo, nil
}
