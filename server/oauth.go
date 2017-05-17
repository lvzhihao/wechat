package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/lvzhihao/wechat/core"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

///connect/oauth2/authorize
func Oauth2Authorize(c *echo.Context) {
	scope := c.Param("scope")
	redirect := c.Param("redirect_uri")
	if scope == "" {
		scope = "snsapi_base"
	}
	url := fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s#wechat_redirect",
		viper.GetString("appid"),
		urLogger.QueryEscape(c.Echo().URL("connect/oauth2/callback")),
		scope,
		urLogger.QueryEscape(redirect),
	)
	Logger.Info("oauth authorize", zap.String("url", url))
	c.Redirect(302, url)
}

////connect/oauth2/callback
func Oauth2Callback(c *echo.Context) {
	//todo 设置安全回调域
	Logger.Sugar().Infof("Request: %v", r)
	code := c.Param("code")
	state := c.Param("state")
	userToken, eerr := Client.GetUserAccessToken(code)
	if eerr != nil {
		Logger.Error("Fetch token error", zap.Any("error", eerr))
		c.Redirect(302, state)
		return
	}
	Logger.Debug("token response", zap.Any("token", userToken))
	userToken.UpdateTime = time.Now()
	collection := Session.DB("").C("user_token")
	_, err := collection.Upsert(bson.M{"openid": userToken.OpenId}, userToken)
	if err != nil {
		Logger.Error("token update mongo err", zap.Error(err))
	}
	if strings.Index(userToken.Scope, "snsapi_userinfo") >= 0 {
		userInfo, eerr := Client.GetUserInfoByToken(userToken)
		if eerr != nil {
			Logger.Error("Fetch userinfo error", zap.Any("error", eerr))
			c.Redirect(302, state)
			return
		}
		Logger.Debug("userinfo", zap.Any("user", userInfo))
		userInfo.UpdateTime = time.Now()
		collection := Session.DB("").C("user_info")
		_, err := collection.Upsert(bson.M{"openid": userInfo.OpenId}, userInfo)
		if err != nil {
			Logger.Error("info update mongo err", zap.Error(err))
		}
	}
	Cache.Set([]byte(code), []byte(userToken.OpenId), 300)
	c.Redirect(302, state+"?code="+code, 302)
	return
}

////connect/oauth2/access_token
func AccessToken(w http.ResponseWriter, r *http.Request) {
	Logger.Sugar().Infof("Request: %v", r)
	code := r.URL.Query().Get("code")
	openid, err := Cache.Get([]byte(code))
	if err != nil {
		ret := core.ClientError{
			ErrCode: -2,
			ErrMsg:  err.Error(),
		}
		b, _ := json.Marshal(ret)
		w.Write(b)
	} else {
		Logger.Debug("code fetch success", zap.String("code", code), zap.String("openid", string(openid)))
		cache.Del([]byte(code))
		collection := s.DB("").C("user_info")
		var userInfo core.UserInfo
		var b []byte
		err := collection.Find(bson.M{"openid": string(openid)}).One(&userInfo)
		if err != nil {
			Logger.Error("info find mongo err", zap.Error(err))
			ret := map[string]string{
				"openid": string(openid),
			}
			b, _ = json.Marshal(ret)
		} else {
			b, _ = json.Marshal(userInfo)
		}
		w.Write(b)
		return
	}
}
