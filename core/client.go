package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var (
	API map[string]string = map[string]string{
		"TOKEN":      "https://api.weixin.qq.com/cgi-bin/token?grant_type=%s&appid=%s&secret=%s",
		"CALLBACKIP": "https://api.weixin.qq.com/cgi-bin/getcallbackip?access_token=%s",
	}
)

type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (t *AccessToken) Token() string {
	return t.AccessToken
	//todo refersh token
}

type ClientError struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (e *ClientError) String() string {
	return fmt.Sprintf("ErrCode: %s; ErrMsg: %s", e.ErrCode, e.ErrMsg)
}

type ClientConfig struct {
	AppId          string        `json:"app_id"`
	AppSecret      string        `json:"app_secret"`
	DefaultTimeout time.Duration `json:"default_timeout"`
}

type Client struct {
	Config  *ClientConfig `json:"config"`
	Token   AccessToken   `json:"token"`
	TokenLk *sync.Mutex
	Error   []ClientError `json:"error"`
}

//todo time.After refresh Tokeno
//todo token write databases

func New(config *ClientConfig) (*Client, *ClientError) {
	if config.AppId == "" {
		return nil, &ClientError{
			ErrCode: 10000,
			ErrMsg:  "缺少AppID",
		}
	}
	if config.AppSecret == "" {
		return nil, &ClientError{
			ErrCode: 10001,
			ErrMsg:  "缺少AppSecret",
		}
	}
	if config.DefaultTimeout <= 0 {
		config.DefaultTimeout = 10 * time.Second
	}
	client := &Client{
		Config: config,
	}
	err := client.RefreshToken()
	if err != nil {
		return nil, err
	} else {
		return client, nil
	}
}

func (c *Client) FetchToken() string {
	return c.Token.Token()
}

func (c *Client) RefreshToken() *ClientError {
	b, err := c.Request(fmt.Sprintf(API["TOKEN"], "client_credential", c.Config.AppId, c.Config.AppSecret), nil)
	if err != nil {
		return err
	}
	eerr := json.Unmarshal(b, &c.Token)
	if eerr == nil && c.Token.Token() != "" {
		return nil
	} else {
		var retErr ClientError
		eerr = json.Unmarshal(b, &retErr)
		if err == nil {
			return &retErr
		} else {
			return &ClientError{ErrCode: -2, ErrMsg: eerr.Error()}
		}
	}
}

func (c *Client) Request(url string, b []byte) ([]byte, *ClientError) {
	client := &http.Client{
		Timeout: c.Config.DefaultTimeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, &ClientError{
			ErrCode: -2,
			ErrMsg:  err.Error(),
		}
	}
	defer resp.Body.Close()
	b, eerr := ioutil.ReadAll(resp.Body)
	if eerr != nil {
		return nil, &ClientError{
			ErrCode: -2,
			ErrMsg:  eerr.Error(),
		}
	}
	return b, nil
}
