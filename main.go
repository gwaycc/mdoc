package main

import (
	"fmt"
	"os"
	"os/signal"

	_ "github.com/gwaycc/mdoc/route"

	"github.com/gwaylib/eweb"
	"github.com/gwaylib/log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var e = eweb.Default()

// Register router
func init() {
	e.Debug = os.Getenv("EWEB_MODE") != "release"

	// middle ware
	e.Use(middleware.Gzip())

	// filter
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			uri := req.URL.Path
			switch uri {
			case "/check": // alive check
				return c.String(200, "1")
			}

			// TODO: auth for login

			// next route
			return next(c)
		}
	})

	// static file
	e.Static("/", "./public")
}

func main() {
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
