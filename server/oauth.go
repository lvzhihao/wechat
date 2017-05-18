package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo"
	"github.com/lvzhihao/wechat/models"
	"go.uber.org/zap"
)

func Oauth2Authorize(ctx echo.Context) error {
	appid := ctx.Get("appid").(string)
	if appid == "" {
		return ctx.NoContent(http.StatusForbidden)
	}
	scope := ctx.QueryParam("scope")
	redirect := ctx.QueryParam("redirect_uri")
	if scope == "" {
		scope = "snsapi_base"
	}
	scheme := ctx.Request().Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}
	url := fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s#wechat_redirect",
		appid,
		url.QueryEscape(scheme+"://"+ctx.Request().Host+ctx.Echo().URL(Oauth2Callback)),
		scope,
		url.QueryEscape(redirect),
	)
	Logger.Info("oauth authorize", zap.String("url", url))
	return ctx.Redirect(http.StatusFound, url)
}

func Oauth2Callback(ctx echo.Context) error {
	//appid := ctx.Get("appid").(string)
	//todo 设置安全回调域
	code := ctx.QueryParam("code")
	state := ctx.QueryParam("state")
	return ctx.Redirect(http.StatusFound, state+"?code="+code)
}

func Oauth2AccessToken(ctx echo.Context) error {
	appid := ctx.Get("appid").(string)
	client, ok := Clients[appid]
	if !ok {
		return ctx.NoContent(http.StatusForbidden)
	}
	code := ctx.QueryParam("code")

	userToken, eerr := client.GetUserAccessToken(code)
	if eerr != nil {
		return ctx.JSON(http.StatusOK, eerr)
	}
	Logger.Debug("token response", zap.Any("token", userToken))
	tokenModel := models.UserAccessToken{
		AppId:       appid,
		OpenId:      userToken.OpenId,
		AccessToken: *userToken,
	}
	err := tokenModel.Upsert(Session)
	if err != nil {
		Logger.Error("token update mongo err", zap.Error(err))
	}
	if strings.Index(userToken.Scope, "snsapi_userinfo") >= 0 {
		userInfo, eerr := client.GetUserInfoByToken(userToken)
		if eerr != nil {
			Logger.Error("Fetch userinfo error", zap.Any("error", eerr))
			return ctx.JSON(http.StatusOK, eerr)
		}
		Logger.Debug("userinfo", zap.Any("user", userInfo))
		infoModel := models.UserInfo{
			AppId:  appid,
			OpenId: userInfo.OpenId,
			Info:   *userInfo,
		}
		err := infoModel.Upsert(Session)
		if err != nil {
			Logger.Error("info update mongo err", zap.Error(err))
		}
		return ctx.JSON(http.StatusOK, userInfo)
	} else {
		ret := map[string]string{
			"openid": userToken.OpenId,
		}
		return ctx.JSON(http.StatusOK, ret)
	}
}
