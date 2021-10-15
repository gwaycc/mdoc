package auth

import (
	"github.com/gwaylib/database"
	"github.com/gwaylib/errors"
)

const (
	USER_KIND_ADMIN  = 1
	USER_KIND_COMMON = 2
)

type UserInfo struct {
	ID       string `db:"id"`
	Passwd   string `db:"passwd"`
	NickName string `db:"nick_name"`
	Kind     int    `db:"kind"`
}

func AddUser(uInfo *UserInfo) error {
	db := GetDB()
	if uInfo.Kind == 0 {
		uInfo.Kind = USER_KIND_COMMON // if the kind not set, fix to common user.
	}
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

func ResetPwd(username, passwd string) error {
	db := GetDB()
	if _, err := db.Exec("UPDATE user_info set passwd=? WHERE id=?", passwd, username); err != nil {
		return errors.As(err, username)
	}
	return nil
}
