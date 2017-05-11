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
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/hashicorp/consul/api"
	"github.com/lvzhihao/wechat/core"
	"github.com/lvzhihao/wechat/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Wechat Proxy Server",
	Long: `Auto Refresh Access Token. For example:

wechat server --app_id=xxxx --app_secret=xxxx`,
	Run: func(cmd *cobra.Command, args []string) {
		//logger
		logger, _ := zap.NewProduction()
		defer logger.Sync() // flushes buffer, if any
		//wechat ulClientlient
		client, eerr := core.New(&core.ClientConfig{
			AppId:          viper.GetString("appid"),
			AppSecret:      viper.GetString("secret"),
			DefaultTimeout: 10 * time.Second,
		})
		//wechat config error, panic
		if eerr != nil {
			logger.Panic("wechat config error", zap.Any("error", eerr))
		}
		logger.Info("Wechat Connecting...", zap.String("token", client.FetchToken()))
		hijack(client, logger)
		health()
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

		go func() {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, os.Kill)
			for {
				<-quit
				if viper.GetString("consul") != "" {
					err := consulClient.Agent().ServiceDeregister(service_id)
					if err != nil {
						logger.Fatal("consul deregister error", zap.Error(err))
					} else {
						logger.Info("Deregister service in consul", zap.String("servcie_id", service_id))
					}
				}
				/*
					ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
					server.Shutdown(ctx)
				*/
				os.Exit(1)
			}
		}()
		if viper.GetString("cert") != "" {
			logger.Fatal("Proxy Stop...", zap.Any("info", http.ListenAndServeTLS(
				viper.GetString("addr"),
				viper.GetString("cert"),
				viper.GetString("key"),
				nil,
			)))
		} else {
			logger.Fatal("Proxy Stop...", zap.Any("info", http.ListenAndServe(
				viper.GetString("addr"),
				nil,
			)))
		}
	},
}

func errResult(code int, msg string) string {
	b, _ := json.Marshal(struct {
		errcode int    `json:"errcode"`
		errmsg  string `json:"errmsg"`
	}{
		code,
		msg,
	})
	return string(b)
}

func health() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
}

func hijack(c *core.Client, l *zap.Logger) {
	sugar := l.Sugar()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requestId := utils.RandStringRunes(40)
		w.Header().Set("X-REQUEST-ID", requestId)

		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, errResult(-2, "not a hijacker"), http.StatusInternalServerError)
			return
		}

		in, _, err := hj.Hijack()
		if err != nil {
			l.Error("Hijack error", zap.Any("url", r.URL), zap.Error(err), zap.String("request_id", requestId))
			http.Error(w, errResult(-2, "hijack error"), http.StatusInternalServerError)
			return
		}
		defer in.Close()

		r.URL.Scheme = "https"
		r.URL.Host = "api.weixin.qq.com:443"
		v := r.URL.Query()
		v.Set("access_token", c.FetchToken())
		r.URL.RawQuery = v.Encode()

		conn, err := net.Dial("tcp", r.URL.Host)
		if err != nil {
			l.Error("Proxy error", zap.Any("url", r.URL), zap.Error(err), zap.String("request_id", requestId))
			http.Error(w, errResult(-2, "hijack error"), http.StatusInternalServerError)
			return
		}
		defer conn.Close()
		out := tls.Client(conn, &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         "api.weixin.qq.com", //must config, see tls.config
		})
		defer out.Close()
		err = r.Write(out)
		if err != nil {
			l.Error("Error copying request", zap.Any("url", r.URL), zap.Error(err), zap.String("request_id", requestId))
			http.Error(w, errResult(-2, "error copying request"), http.StatusInternalServerError)
			return
		}

		errc := make(chan error, 2)
		cp := func(dst io.Writer, src io.Reader) {
			_, err := io.Copy(dst, src)
			errc <- err
		}

		sugar.Infof("Request: %v", r)

		go cp(out, in)
		go cp(in, out)
		err = <-errc
		if err != nil && err != io.EOF {
			l.Error("proxy error", zap.Any("url", r.URL), zap.Error(err), zap.String("request_id", requestId))
		}
	})
}

func init() {
	RootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	serverCmd.Flags().String("appid", "", "第三方用户唯一凭证")
	serverCmd.Flags().String("secret", "", "第三方用户唯一凭证密钥")
	serverCmd.Flags().String("addr", "127.0.0.1:9099", "代理监听地址")
	serverCmd.Flags().String("cert", "", "ssl证书")
	serverCmd.Flags().String("key", "", "ssl证书")
	serverCmd.Flags().String("consul", "", "consul api")
	serverCmd.Flags().String("token", "", "consul acl token")
	viper.BindPFlag("appid", serverCmd.Flags().Lookup("appid"))
	viper.BindPFlag("secret", serverCmd.Flags().Lookup("secret"))
	viper.BindPFlag("addr", serverCmd.Flags().Lookup("addr"))
	viper.BindPFlag("cert", serverCmd.Flags().Lookup("cert"))
	viper.BindPFlag("key", serverCmd.Flags().Lookup("key"))
	viper.BindPFlag("consul", serverCmd.Flags().Lookup("consul"))
	viper.BindPFlag("token", serverCmd.Flags().Lookup("token"))
}
