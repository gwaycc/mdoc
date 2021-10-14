package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/gwaycc/mdoc/route"
	"github.com/gwaycc/mdoc/tools/auth"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/eweb"
	"github.com/gwaylib/log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var e = eweb.Default()

var (
	authFlag   = flag.Bool("auth-mode", true, "run the authentication mode")
	dbFileFlag = flag.String("db-file", "./data/mdoc.db", "where db file to storage")
)

func authPasswd(user, realm string) string {
	uInfo, err := auth.GetUser(user)
	if err != nil {
		if errors.ErrNoData.Equal(err) {
			return ""
		}
		log.Warn(errors.As(err, user, realm))
		return ""
	}
	return uInfo.Passwd
}

// Register router
func init() {
	e.Debug = os.Getenv("EWEB_MODE") != "release"

	auth.InitDB(*dbFileFlag)
	digestLogin := auth.NewDigestAuth("lib10", true, authPasswd)

	// digest auth

	// render
	e.Renderer = eweb.GlobTemplate("./public/*.html")

	// middle ware
	e.Use(middleware.Gzip())

	// filter
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			uri := req.URL.Path

			route.DumpReq(req)
			switch uri {
			case "/check": // alive check
				return c.String(200, "1")
			case "/favicon.ico":
				// continue
			default:
				if *authFlag {
					// login check
					username, err := digestLogin.CheckAuth(c.Response().Writer, req)
					if err != nil {
						if auth.ErrNeedPwd.Equal(err) {
							log.Info(errors.As(err))
						}
						return nil
					}
					_ = username
				}
			}

			// next route
			return next(c)
		}
	})

	// static file
	e.Static("/", "./public")
}

func main() {
	flag.Parse()

	// Start server
	go func() {
		httpAddr := ":8080"
		log.Infof("Http listen : %s", httpAddr)
		log.Exit(2, e.Start(httpAddr))
	}()

	// exit event
	fmt.Println("[ctrl+c to exit]")
	end := make(chan os.Signal, 2)
	signal.Notify(end, os.Interrupt, os.Kill)
	<-end
}
