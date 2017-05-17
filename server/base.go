package server

import (
	"encoding/json"
	"net/http"

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

	AppId        string
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

func Health(c echo.Context) error {
	return c.HTML(http.StatusOK, "ok")
}
