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
	"net/http/httputil"
	"time"

	"go.uber.org/zap"

	"github.com/lvzhihao/wechat/core"
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
		//wechat client
		client, err := core.New(&core.ClientConfig{
			AppId:          viper.GetString("app_id"),
			AppSecret:      viper.GetString("app_secret"),
			DefaultTimeout: 10 * time.Second,
		})
		//wechat config error, panic
		if err != nil {
			logger.Panic("wechat config error", zap.Any("error", err))
		}
		logger.Info("Wechat Connecting...", zap.String("token", client.FetchToken()))
		hijack(client, logger)
		logger.Info("Proxy Running...", zap.String("addr", ":9099"))
		logger.Fatal("Proxy Stop...", zap.Any("info", http.ListenAndServeTLS(":9099", "./pem/server.cert", "./pem/server.key", nil)))
	},
}

func errResult(code int, msg string) string {
	b, _ := json.Marshal(struct {
		errcode int
		errmsg  string
	}{
		code,
		msg,
	})
	return string(b)
}

func hijack(c *core.Client, l *zap.Logger) {
	sugar := l.Sugar()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, errResult(-2, "not a hijacker"), http.StatusInternalServerError)
			return
		}

		in, _, err := hj.Hijack()
		if err != nil {
			l.Error("Hijack error", zap.Any("url", r.URL), zap.Error(err))
			http.Error(w, errResult(-2, "hijack error"), http.StatusInternalServerError)
			return
		}
		defer in.Close()

		r.URL.Scheme = "https"
		r.URL.Host = "api.weixin.qq.com:443"
		v := r.URL.Query()
		v.Set("access_token", c.FetchToken())
		r.URL.RawQuery = v.Encode()

		dial, err := net.Dial("tcp", r.URL.Host)
		tls_conn := tls.Client(dial, &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         "api.weixin.qq.com", //must config, see tls.config
		})
		out, _ := httputil.NewClientConn(tls_conn, nil).Hijack()
		err = r.Write(out)
		if err != nil {
			l.Error("Error copying request", zap.Any("url", r.URL), zap.Error(err))
			http.Error(w, errResult(-2, "error copying request"), http.StatusInternalServerError)
			return
		}
		defer out.Close()

		errc := make(chan error, 2)
		cp := func(dst io.Writer, src io.Reader) {
			_, err := io.Copy(dst, src)
			errc <- err
		}

		sugar.Debugf("Request: %v", r)

		go cp(out, in)
		go cp(in, out)
		err = <-errc
		if err != nil && err != io.EOF {
			l.Error("proxy error", zap.Any("url", r.URL), zap.Error(err))
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
	viper.Set("app_id", serverCmd.Flags().String("app_id", "", "第三方用户唯一凭证"))
	viper.Set("app_secret", serverCmd.Flags().String("app_secret", "", "第三方用户唯一凭证密钥"))
}
