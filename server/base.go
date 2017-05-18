package server

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo"
	"github.com/lvzhihao/wechat/core"
	"go.uber.org/zap"
	mgo "gopkg.in/mgo.v2"
)

var (
	Logger  *zap.Logger
	Session *mgo.Session

	Clients       map[string]*core.Client
	ReceiveTokens map[string]string
	CallbckUrls   []string
)

func init() {
	Clients = make(map[string]*core.Client, 0)
	ReceiveTokens = make(map[string]string, 0)
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

func Health(ctx echo.Context) error {
	return ctx.HTML(http.StatusOK, "ok")
}
