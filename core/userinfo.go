package core

import (
	"encoding/json"
	"fmt"
)

type UserAccessToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"open_id"`
	Scope        string `json:"scope"`
}

type UserInfo struct {
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
	var res map[string]interface{}
	err := json.Unmarshal(b, &res)
	if err != nil {
		return nil, &ClientError{ErrCode: -2, ErrMsg: err.Error()}
	}
	var token UserAccessToken
	json.Unmarshal(b, &token)
	return &token, nil
}

func (c *Client) GetUserInfoByToken(token *UserAccessToken) (*UserInfo, *ClientError) {
	//todo
	return nil, nil
}
