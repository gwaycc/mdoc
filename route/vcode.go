package route

import (
	"fmt"
	"time"

	"github.com/dchest/captcha"
	"github.com/gwaylib/eweb"
	"github.com/labstack/echo"
)

func init() {
	e := eweb.Default()
	e.GET("/api/vcode/:id", echo.WrapHandler(captcha.Server(130, 53)))
	e.POST("/api/vcode", GetVCode)
}

func GetVCode(c echo.Context) error {
	lastId := FormValue(c, "id")

	// unused
	if len(lastId) > 0 && captcha.Reload(lastId) {
		return c.JSON(200, eweb.H{
			"id":  lastId,
			"img": fmt.Sprintf("/api/vcode/%s.png?rand=%d", lastId, time.Now().Unix()),
		})
	}

	// new code
	id := captcha.NewLen(4)
	return c.JSON(200, eweb.H{
		"id":  id,
		"img": fmt.Sprintf("/api/vcode/%s.png", id),
	})
}
