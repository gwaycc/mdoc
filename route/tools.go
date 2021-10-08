package route

import (
	"fmt"
	"time"

	"net/http"
	"net/http/httputil"

	"github.com/gwaylib/errors"
	"github.com/labstack/echo"
)

func DumpReq(req *http.Request) {
	data, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(data))
	}
}

func FormValue(c echo.Context, key string) string {
	req := c.Request()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req.FormValue(key)
}

func ParseTime(timeStr string) (time.Time, error) {
	t, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return t, errors.As(err)
	}
	return t, nil
}
