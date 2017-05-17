package server

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"

	"github.com/lvzhihao/wechat/core"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func receive(w http.ResponseWriter, r *http.Request) {
	if core.ReceiveMsgCheckSign(viper.GetString("token"), r) {
		q := r.URL.Query()
		if q.Get("echostr") != "" {
			w.Write([]byte(q.Get("echostr")))
		} else {
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				l.Error("request body empty", zap.Error(err))
				http.NotFound(w, r)
			} else {
				l.Debug("request body ", zap.String("body", string(data)))
				var msg core.Msg
				err := xml.Unmarshal(data, &msg)
				if err != nil {
					l.Error("request body except", zap.Error(err))
					http.NotFound(w, r)
				} else {
					l.Debug("xml content", zap.Any("xml", msg))
					w.Write(nil)
					//todo
					/*
							ret := &core.RetTextMsg{RetMsgComm: core.RetMsgComm{
								ToUserName:   msg.FromUserName,
								FromUserName: msg.ToUserName,
								CreateTime:   int(time.Now().Unix()),
								MsgType:      "text",
							}, Content: "replay test"}
							b, err := xml.Marshal(ret)
						if err != nil {
							l.Error("msg reply error", zap.Error(err))
							//retry todo
							w.Write([]byte("success"))
						} else {
							l.Debug("msg reply", zap.String("xml", string(b)))
							w.Write(b)
						}
					*/
				}
			}
		}
	} else {
		http.NotFound(w, r)
	}
}
