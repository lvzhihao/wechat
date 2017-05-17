package server

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"

	"github.com/lvzhihao/wechat/utils"
	"go.uber.org/zap"
)

func proxy(w http.ResponseWriter, r *http.Request) {
	sugar := Logger.Sugar()
	requestId := utils.RandStringRunes(40)
	w.Header().Set("X-REQUEST-ID", requestId)

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, errResult(-2, "not a hijacker"), http.StatusInternalServerError)
		return
	}

	in, _, err := hj.Hijack()
	if err != nil {
		Logger.Error("Hijack error", zap.Any("url", r.URL), zap.Error(err), zap.String("request_id", requestId))
		http.Error(w, errResult(-2, "hijack error"), http.StatusInternalServerError)
		return
	}
	defer in.Close()

	r.URL.Scheme = "https"
	r.URL.Host = "api.weixin.qq.com:443"
	v := r.URL.Query()
	v.Set("access_token", Client.FetchToken())
	r.URL.RawQuery = v.Encode()

	conn, err := net.Dial("tcp", r.URL.Host)
	if err != nil {
		Logger.Error("Proxy error", zap.Any("url", r.URL), zap.Error(err), zap.String("request_id", requestId))
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
		Logger.Error("Error copying request", zap.Any("url", r.URL), zap.Error(err), zap.String("request_id", requestId))
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
		Logger.Error("proxy error", zap.Any("url", r.URL), zap.Error(err), zap.String("request_id", requestId))
	}
}
