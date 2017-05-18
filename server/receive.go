package server

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"
	"github.com/lvzhihao/wechat/core"
	"go.uber.org/zap"
)

func Receive(ctx echo.Context) error {
	appid := ctx.Get("appid").(string)
	token, ok := ReceiveTokens[appid]
	if !ok {
		return ctx.NoContent(http.StatusForbidden)
	}
	if !core.ReceiveMsgCheckSign(token, ctx.Request()) {
		return ctx.NoContent(http.StatusForbidden)
	}
	if ctx.QueryParam("echostr") != "" {
		return ctx.String(http.StatusOK, ctx.QueryParam("echostr"))
	} else {
		data, err := ioutil.ReadAll(ctx.Request().Body)
		if err != nil {
			Logger.Error("request body empty", zap.Error(err))
			return ctx.NoContent(http.StatusNotFound)
		} else {
			Logger.Debug("request body ", zap.String("body", string(data)))
			var msg core.Msg
			err := xml.Unmarshal(data, &msg)
			if err != nil {
				Logger.Error("request body except", zap.Error(err))
				return ctx.NoContent(http.StatusNotFound)
			} else {
				Logger.Debug("xml content", zap.Any("xml", msg))
				return ctx.String(http.StatusOK, "ignore")
				//todo
				//ret := &core.RetTextMsg{RetMsgComm: core.RetMsgComm{
				//	ToUserName:   msg.FromUserName,
				//	FromUserName: msg.ToUserName,
				//	CreateTime:   int(time.Now().Unix()),
				//	MsgType:      "text",
				//}, Content: "replay test"}
				//b, err := xml.Marshal(ret)
				//if err != nil {
				//	l.Error("msg reply error", zap.Error(err))
				//	//retry todo
				//	w.Write([]byte("success"))
				//} else {
				//	l.Debug("msg reply", zap.String("xml", string(b)))
				//	w.Write(b)
				//}
			}
		}
	}
}
