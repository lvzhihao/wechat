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
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/lvzhihao/wechat/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := core.New(&core.ClientConfig{
			AppId:          viper.GetString("app_id"),
			AppSecret:      viper.GetString("app_secret"),
			DefaultTimeout: 10 * time.Second,
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Connecting..., token is %s\n", client.FetchToken())
		hijack(client)
		log.Printf("Server run at %s\n", ":9099")
		log.Fatal(http.ListenAndServeTLS(":9099", "./pem/server.cert", "./pem/server.key", nil))
		//log.Fatal(http.ListenAndServe(":9099", nil))
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

func hijack(c *core.Client) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, errResult(-2, "not a hijacker"), http.StatusInternalServerError)
			return
		}

		in, _, err := hj.Hijack()
		if err != nil {
			log.Printf("[ERROR] Hijack error for %s. %s", r.URL, err)
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
			ServerName:         "api.weixin.qq.com",
		})
		out, _ := httputil.NewClientConn(tls_conn, nil).Hijack()
		err = r.Write(out)
		if err != nil {
			log.Printf("[ERROR] Error copying request for %s. %s", r.URL, err)
			http.Error(w, errResult(-2, "error copying request"), http.StatusInternalServerError)
			return
		}
		defer out.Close()

		errc := make(chan error, 2)
		cp := func(dst io.Writer, src io.Reader) {
			_, err := io.Copy(dst, src)
			errc <- err
		}

		log.Printf("[DEBUG] %v\n", r)

		go cp(out, in)
		go cp(in, out)
		err = <-errc
		if err != nil && err != io.EOF {
			log.Printf("[INFO] WS error for %s. %s", r.URL, err)
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
