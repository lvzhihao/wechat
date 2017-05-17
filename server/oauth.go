package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/lvzhihao/wechat/core"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

func Oauth2Authorize(c echo.Context) error {
	scope := c.QueryParam("scope")
	redirect := c.QueryParam("redirect_uri")
	if scope == "" {
		scope = "snsapi_base"
	}
	scheme := c.Request().Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}
	url := fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s#wechat_redirect",
		AppId,
		url.QueryEscape(scheme+"://"+c.Request().Host+"/"+c.Echo().URL(Oauth2Callback)),
		scope,
		url.QueryEscape(redirect),
	)
	Logger.Info("oauth authorize", zap.String("url", url))
	return c.Redirect(http.StatusFound, url)
}

func Oauth2Callback(c echo.Context) error {
	//todo 设置安全回调域
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	userToken, eerr := Client.GetUserAccessToken(code)
	if eerr != nil {
		Logger.Error("Fetch token error", zap.Any("error", eerr))
		return c.Redirect(http.StatusFound, state)
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
			return c.Redirect(http.StatusFound, state)
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
	return c.Redirect(http.StatusFound, state+"?code="+code)
}

func Oauth2AccessToken(c echo.Context) error {
	code := c.QueryParam("code")
	openid, err := Cache.Get([]byte(code))
	if err != nil {
		ret := core.ClientError{
			ErrCode: -2,
			ErrMsg:  err.Error(),
		}
		return c.JSON(http.StatusOK, ret)
	} else {
		Logger.Debug("code fetch success", zap.String("code", code), zap.String("openid", string(openid)))
		Cache.Del([]byte(code))
		collection := Session.DB("").C("user_info")
		var userInfo core.UserInfo
		err := collection.Find(bson.M{"openid": string(openid)}).One(&userInfo)
		if err != nil {
			Logger.Error("info find mongo err", zap.Error(err))
			ret := map[string]string{
				"openid": string(openid),
			}
			return c.JSON(http.StatusOK, ret)
		} else {
			return c.JSON(http.StatusOK, userInfo)
		}
	}
}
