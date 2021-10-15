package route

import (
	"github.com/gwaycc/mdoc/tools/auth"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/eweb"
	"github.com/gwaylib/log"
	"github.com/labstack/echo"
)

func init() {
	e := eweb.Default()
	e.POST("/user/add", UserAdd)
	e.POST("/user/pwd/reset", UserPwdReset)
}

func UserAdd(c echo.Context) error {
	username := FormValue(c, "username")
	passwd := FormValue(c, "passwd")
	nickName := FormValue(c, "nickname")

	if _, err := auth.GetUser(username); err != nil {
		if !errors.ErrNoData.Equal(err) {
			log.Warn(errors.As(err))
			return c.String(500, "System interval error")
		}
		// pass
	} else {
		return c.String(403, "User already exist.")
	}

	if err := auth.AddUser(&auth.UserInfo{
		ID:       username,
		Passwd:   passwd,
		NickName: nickName,
	}); err != nil {
		log.Warn(errors.As(err))
		return c.String(500, "System interval error")
	}

	return c.String(200, "OK")
}

func UserPwdReset(c echo.Context) error {
	username := FormValue(c, "username")
	passwd := FormValue(c, "passwd")
	if err := auth.ResetPwd(username, passwd); err != nil {
		log.Warn(errors.As(err))
		return c.String(500, "System interval error")
	}
	auth.DelAuthCache(username)
	return c.String(200, "OK")
}
