package server

import (
	"encoding/json"

	"github.com/coocood/freecache"
	"github.com/labstack/echo"
	"github.com/lvzhihao/wechat/core"
	"go.uber.org/zap"
	mgo "gopkg.in/mgo.v2"
)

var (
	Logger  *zap.Logger
	Client  *core.Client
	Cache   *freecache.Cache
	Session *mgo.Session

	ReceiveToken string
)

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

func health(c echo.Context) {
	return c.String("ok")
}
