// Copyright © 2017 edwin <edwin.lzh@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"

	"go.uber.org/zap"

	"github.com/hashicorp/consul/api"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/lvzhihao/wechat/core"
	"github.com/lvzhihao/wechat/models"
	"github.com/lvzhihao/wechat/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ServerKey struct {
	Key    string
	Secret string
	AppId  string
}

var searchKeys map[string]*ServerKey

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Wechat Proxy Server",
	Long: `Auto Refresh Access Token. For example:

wechat server --app_id=xxxx --app_secret=xxxx`,
	Run: func(cmd *cobra.Command, args []string) {
		// logger
		var logger *zap.Logger
		if viper.GetBool("debug") {
			logger, _ = zap.NewDevelopment()
		} else {
			logger, _ = zap.NewProduction()
		}
		defer logger.Sync() // flushes buffer, if any

		// mongo
		session, err := mgo.Dial(viper.GetString("mongo"))
		if err != nil {
			logger.Panic("mongo config error", zap.Error(err))
		}
		defer session.Close()

		// ensure mongo index
		models.UserAccessTokenEnsureIndex(session)
		models.UserInfoEnsureIndex(session)

		// wechat Client
		lists, err := models.ListWechatConfig(session)
		if err != nil {
			log.Printf("Load WechatConfig Error: %s\n", err)
		} else {
			for _, config := range lists {
				client, eerr := core.New(&core.ClientConfig{
					AppId:          config.AppId,
					AppSecret:      config.AppSecret,
					DefaultTimeout: 10 * time.Second,
				})
				if eerr != nil {
					logger.Error("wechat config error", zap.Any("error", eerr))
				} else {
					server.Clients[config.AppId] = client
					server.ReceiveTokens[config.AppId] = config.ReceiveToken
					searchKeys[config.Detail.Key] = &ServerKey{
						Key:    config.Detail.Key,
						Secret: config.Detail.Secret,
						AppId:  config.AppId,
					}
				}
			}
		}

		// consul register
		service_id := "wproxy-" + viper.GetString("addr")
		var consulClient *api.Client
		if viper.GetString("consul") != "" {
			var err error
			consulClient, err = api.NewClient(&api.Config{
				Address: viper.GetString("consul"),
				Scheme:  "http",
				Token:   viper.GetString("token"),
			})
			if err != nil {
				logger.Panic("consul config error", zap.Error(err))
			}
			checkScheme := "http"
			if viper.GetString("cert") != "" {
				checkScheme = "https"
			}
			addrList := strings.Split(viper.GetString("addr"), ":")
			host := addrList[0]
			port, _ := strconv.Atoi(addrList[1])
			err = consulClient.Agent().ServiceRegister(&api.AgentServiceRegistration{
				ID:      service_id,
				Name:    "wproxy-" + strconv.Itoa(port),
				Port:    port,
				Address: host,
				Tags:    []string{"v1", "proxy"},
				Check: &api.AgentServiceCheck{
					HTTP:     checkScheme + "://" + viper.GetString("addr") + "/health",
					Interval: "1s",
					Timeout:  "1s",
				},
			})
			if err != nil {
				logger.Panic("consul register error", zap.Error(err))
			} else {
				logger.Info("Register service in consul", zap.String("servcie_id", service_id))
			}
		}
		logger.Info("Proxy Running...", zap.String("addr", viper.GetString("addr")))
		defer func() {
			if viper.GetString("consul") != "" {
				err := consulClient.Agent().ServiceDeregister(service_id)
				if err != nil {
					logger.Fatal("consul deregister error", zap.Error(err))
				} else {
					logger.Info("Deregister service in consul", zap.String("servcie_id", service_id))
				}
			}
		}()

		server.Session = session
		server.Logger = logger

		// echo server
		e := echo.New()
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())

		// health check
		e.Any("/health", server.Health)

		// check searchKeys
		e.Any("/*", server.Proxy, CheckKeys) // api proxy
		e.GET("/connect/oauth2/authorize", server.Oauth2Authorize, CheckKeys)
		e.GET("/connect/oauth2/access_token", server.Oauth2AccessToken, CheckKeys)

		// no check todo 安全回调域
		e.GET("/connect/oauth2/callback", server.Oauth2Callback)

		// receive
		e.Any("/receive/:key", server.Receive, ChoiceKeys)

		// start
		go func() {
			if viper.GetString("cert") != "" {
				e.Logger.Fatal(e.StartTLS(
					viper.GetString("addr"),
					viper.GetString("cert"),
					viper.GetString("key"),
				))
			} else {
				e.Logger.Fatal(e.Start(viper.GetString("addr")))
			}
		}()

		// graceful shutdown
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt, os.Kill)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Fatal(err)
		}
	},
}

func ChoiceKeys(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		key := c.Param("key")
		o, ok := searchKeys[key]
		if !ok {
			return c.NoContent(http.StatusNotFound)
		}
		c.Set("appid", o.AppId)
		return next(c)
	}
}

func CheckKeys(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		key := c.QueryParam("key")
		secret := c.QueryParam("secret")
		o, ok := searchKeys[key]
		if !ok {
			return c.NoContent(http.StatusNotFound)
		}
		if o.Secret != secret {
			return c.NoContent(http.StatusNotFound)
		}
		c.Set("appid", o.AppId)
		return next(c)
	}
}

func init() {
	RootCmd.AddCommand(serverCmd)

	searchKeys = make(map[string]*ServerKey, 0)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	serverCmd.Flags().String("addr", "127.0.0.1:9099", "代理监听地址")
	serverCmd.Flags().String("cert", "", "ssl证书")
	serverCmd.Flags().String("key", "", "ssl证书")
	serverCmd.Flags().String("consul", "", "consul api")
	serverCmd.Flags().String("consul_token", "", "consul acl token")
	serverCmd.Flags().Bool("debug", false, "display debug log")
	serverCmd.Flags().String("mongo", "mongodb://127.0.0.1/wechat", "mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]")
	viper.BindPFlag("addr", serverCmd.Flags().Lookup("addr"))
	viper.BindPFlag("cert", serverCmd.Flags().Lookup("cert"))
	viper.BindPFlag("key", serverCmd.Flags().Lookup("key"))
	viper.BindPFlag("consul", serverCmd.Flags().Lookup("consul"))
	viper.BindPFlag("consul_token", serverCmd.Flags().Lookup("consul_token"))
	viper.BindPFlag("debug", serverCmd.Flags().Lookup("debug"))
	viper.BindPFlag("mongo", serverCmd.Flags().Lookup("mongo"))
}
