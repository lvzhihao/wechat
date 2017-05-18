package server

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"

	"github.com/labstack/echo"
	"github.com/lvzhihao/wechat/utils"
	"go.uber.org/zap"
)

func Proxy(ctx echo.Context) error {
	w := ctx.Response().Writer
	r := ctx.Request()
	appid := ctx.Get("appid").(string)
	client, ok := Clients[appid]
	if !ok {
		return ctx.NoContent(http.StatusForbidden)
	}

	sugar := Logger.Sugar()
	requestId := utils.RandStringRunes(40)

	hj, ok := w.(http.Hijacker)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, errResult(-2, "not a hijacker"))
	}

	in, _, err := hj.Hijack()
	if err != nil {
		Logger.Error("Hijack error", zap.Any("url", r.URL), zap.Error(err), zap.String("request_id", requestId))
		return ctx.JSON(http.StatusInternalServerError, errResult(-2, "hijack error"))
	}
	defer in.Close()

	r.URL.Scheme = "https"
	r.URL.Host = "api.weixin.qq.com:443"
	v := r.URL.Query()
	v.Set("access_token", client.FetchToken())
	r.URL.RawQuery = v.Encode()

	conn, err := net.Dial("tcp", r.URL.Host)
	if err != nil {
		Logger.Error("Proxy error", zap.Any("url", r.URL), zap.Error(err), zap.String("request_id", requestId))
		return ctx.JSON(http.StatusInternalServerError, errResult(-2, "hijack error"))
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
		return ctx.JSON(http.StatusInternalServerError, errResult(-2, "error copying request"))
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
	return nil
}
