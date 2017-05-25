package core

import (
	"encoding/json"
	"fmt"
)

type UserAccessToken struct {
	AccessToken  string `bson:"access_token" json:"access_token"`
	ExpiresIn    int64  `bson:"expires_in" json:"expires_in"`
	RefreshToken string `bson:"refresh_token" json:"refresh_token"`
	OpenId       string `bson:"openid" json:"openid"`
	Scope        string `bson:"scope" json:"scope"`
}

type UserInfo struct {
	OpenId     string   `bson:"openid" json:"openid"`
	NickName   string   `bson:"nickname" json:"nickname"`
	Sex        int64    `bson:"sex" json:"sex"`
	Province   string   `bson:"province" json:"province"`
	City       string   `bson:"city" json:"city"`
	Country    string   `bson:"country" json:"country"`
	HeadImgUrl string   `bson:"headimgurl" json:"headimgurl"`
	Privilege  []string `bson:"privilege" json:"privilege"`
	UnionId    string   `bson:"unionid" json:"unionid"`
}

func (c *Client) RefreshUserAccessToken(refresh_token string) (*UserAccessToken, *ClientError) {
	b, eerr := c.Request(fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/refresh_token?appid=%s&grant_type=refresh_token&refresh_token=%s",
		c.config.AppId,
		refresh_token,
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
