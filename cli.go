package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/gwaycc/mdoc/route"
	"github.com/gwaycc/mdoc/tools/auth"
	"github.com/gwaycc/mdoc/tools/repo"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/eweb"
	"github.com/gwaylib/log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/urfave/cli/v2"
)

// resgister daemon
func init() {
	app.Register("daemon",
		&cli.Command{
			Name: "daemon",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "dump",
					Value: false,
					Usage: "dump the request header",
				},
				&cli.BoolFlag{
					Name:  "auth-mode",
					Value: true,
					Usage: "run the authentication mode",
				},
				&cli.StringFlag{
					Name:  "listen",
					Value: ":8080",
					Usage: "http listen address",
				},
			},
			Action: func(cctx *cli.Context) error {
				ctx := cctx.Context
				_ = ctx

				authMode := cctx.Bool("auth-mode")
				listenAddr := cctx.String("listen")
				repoDir := repo.ExpandPath(cctx.String("repo"))

				// digest auth
				auth.InitDB(filepath.Join(repoDir, "data", "mdoc.db"))
				authPasswd := func(user, realm string) string {
					pwd, ok := auth.GetAuthCache(user)
					if ok {
						return pwd
					}

					uInfo, err := auth.GetUser(user)
					if err != nil {
						if errors.ErrNoData.Equal(err) {
							return ""
						}
						log.Warn(errors.As(err, user, realm))
						return ""
					}
					auth.UpdateAuthCache(user, uInfo.Passwd)
					return uInfo.Passwd
				}
				digestLogin := auth.NewDigestAuth(auth.REALM, false, authPasswd)

				// web server
				var e = eweb.Default()
				e.Debug = os.Getenv("EWEB_MODE") != "release"
				e.Renderer = eweb.GlobTemplate(filepath.Join(repoDir, "public", "*.html"))

				// middle ware
				e.Use(middleware.Gzip())

				// filter
				dump := cctx.Bool("dump")
				e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
					return func(c echo.Context) error {
						req := c.Request()
						uri := req.URL.Path
						if dump {
							route.DumpReq(req)
						}

						switch uri {
						case "/check": // alive check
							return c.String(200, "1")
						case "/favicon.ico":
							// continue
						default:
							if authMode {
								// login check
								username, err := digestLogin.CheckAuth(req)
								switch {
								case auth.ErrNeedLogin.Equal(err):
									digestLogin.RequireAuth(c.Response().Writer, req)
									return nil
								case auth.ErrNeedPwd.Equal(err):
									log.Info(errors.As(err))
									digestLogin.RequireAuth(c.Response().Writer, req)
									return nil
								case auth.ErrReject.Equal(err):
									return c.String(403, auth.ErrReject.Code())
								default:
									if err != nil {
										log.Warn(errors.As(err))
										return c.String(500, "unknow error")
									}

									// login success
								}
								_ = username
							}
						}

						// next route
						return next(c)
					}
				})

				// static file
				e.Static("/", filepath.Join(repoDir, "public"))

				// Start server
				go func() {
					log.Exit(2, e.Start(listenAddr))
				}()

				log.Infof("Http listen: %s, [ctrl+c to exit]", listenAddr)
				// exit event
				end := make(chan os.Signal, 2)
				signal.Notify(end, os.Interrupt, os.Kill)
				<-end

				return nil
			},
		},
	)
}

// resgister user tool
func init() {
	app.Register("user",
		&cli.Command{
			Name: "user",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "url",
					Value: "http://127.0.0.1:8080",
					Usage: "server url",
				},
				&cli.StringFlag{
					Name:  "admin-user",
					Value: "admin",
					Usage: "input the admin user",
				},
				&cli.StringFlag{
					Name:  "admin-pwd",
					Value: "",
					Usage: "input the admin's password",
				},
			},
			Subcommands: []*cli.Command{
				&cli.Command{
					Name:  "add",
					Usage: "add a new user",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "username",
							Value: "",
							Usage: "input the username",
						},
						&cli.StringFlag{
							Name:  "passwd",
							Value: "",
							Usage: "input the password",
						},
						&cli.StringFlag{
							Name:  "nickname",
							Value: "",
							Usage: "input the nickname",
						},
					},
					Action: func(cctx *cli.Context) error {
						ctx := cctx.Context
						_ = ctx

						username := cctx.String("username")
						passwd := cctx.String("passwd")
						nickName := cctx.String("nickname")

						params := url.Values{
							"username": {username},
							"passwd":   {auth.HashPasswd(username, auth.REALM, passwd)},
							"nickname": {nickName},
						}
						if err := auth.AuthReq(
							cctx.String("url"), "/user/add",
							cctx.String("admin-user"), cctx.String("admin-pwd"),
							params); err != nil {
							return errors.As(err)
						}
						fmt.Println("add user success")
						return nil

					},
				},
				&cli.Command{
					Name:  "reset",
					Usage: "reset the user's password",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "username",
							Value: "",
							Usage: "input the username",
						},
						&cli.StringFlag{
							Name:  "passwd",
							Value: "",
							Usage: "input the new password",
						},
					},
					Action: func(cctx *cli.Context) error {
						ctx := cctx.Context
						_ = ctx

						username := cctx.String("username")
						passwd := cctx.String("passwd")
						params := url.Values{
							"username": {username},
							"passwd":   {auth.HashPasswd(username, auth.REALM, passwd)},
						}
						if err := auth.AuthReq(
							cctx.String("url"), "/user/pwd/reset",
							cctx.String("admin-user"), cctx.String("admin-pwd"),
							params); err != nil {
							return errors.As(err)
						}
						fmt.Println("change password success")
						return nil
					},
				},
			},
		},
	)
}
