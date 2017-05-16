package core

import (
	"encoding/json"
	"fmt"
)

type UserAccessToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"openid"`
	Scope        string `json:"scope"`
}

type UserInfo struct {
	OpenId     string   `json:"openid"`
	NickName   string   `json:"nickname"`
	Sex        int64    `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgUrl string   `json:"headimgurl"`
	Privilege  []string `json:"privielge"`
	UnionId    string   `json:"unionid"`
}

func (c *Client) GetUserAccessToken(code string) (*UserAccessToken, *ClientError) {
	b, eerr := c.Request(fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		c.config.AppId,
		c.config.AppSecret,
		code,
	))
	if eerr != nil {
		return nil, eerr
	}
	var token UserAccessToken
	err := json.Unmarshal(b, &token)
	if err != nil {
		return nil, &ClientError{ErrCode: -2, ErrMsg: err.Error()}
	} else {
		return &token, nil
	}
}

func (c *Client) GetUserInfoByToken(token *UserAccessToken) (*UserInfo, *ClientError) {
	b, eerr := c.Request(fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN",
		token.AccessToken,
		token.OpenId,
	))
	if eerr != nil {
		return nil, eerr
	}
	var info UserInfo
	err := json.Unmarshal(b, &info)
	if err != nil {
		return nil, &ClientError{ErrCode: -2, ErrMsg: err.Error()}
	} else {
		return &info, nil
	}
}
