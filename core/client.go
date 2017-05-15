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
	BaseApis map[string]string = map[string]string{
		"TOKEN":      "https://api.weixin.qq.com/cgi-bin/token?grant_type=%s&appid=%s&secret=%s",
		"CALLBACKIP": "https://api.weixin.qq.com/cgi-bin/getcallbackip?access_token=%s",
	}
)

type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type ClientError struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (e *ClientError) String() string {
	return fmt.Sprintf("ErrCode: %d; ErrMsg: %s", e.ErrCode, e.ErrMsg)
}

type ClientConfig struct {
	AppId          string        `json:"app_id"`
	AppSecret      string        `json:"app_secret"`
	DefaultTimeout time.Duration `json:"default_timeout"`
}

type Client struct {
	config *ClientConfig `json:"config"`
	token  AccessToken   `json:"token"`
	lock   sync.Mutex    `json:"-"`
	Error  []ClientError `json:"error"`
}

//todo time.After refresh Tokeno
//todo token write databases

func New(config *ClientConfig) (*Client, *ClientError) {
	if config.AppId == "" {
		return nil, &ClientError{
			ErrCode: -2,
			ErrMsg:  "缺少AppID",
		}
	}
	if config.AppSecret == "" {
		return nil, &ClientError{
			ErrCode: -2,
			ErrMsg:  "缺少AppSecret",
		}
	}
	if config.DefaultTimeout <= 0 {
		config.DefaultTimeout = 10 * time.Second
	}
	client := &Client{
		config: config,
	}
	err := client.RefreshToken()
	if err != nil {
		return nil, err
	} else {
		return client, nil
	}
}

func (c *Client) Token() string {
	return c.FetchToken()
}

func (c *Client) FetchToken() string {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.token.ExpiresIn > time.Now().Unix()-10 {
		c.RefreshToken()
	}
	return c.token.AccessToken
}

func (c *Client) RefreshToken() *ClientError {
	b, err := c.Request(fmt.Sprintf(BaseApis["TOKEN"], "client_credential", c.config.AppId, c.config.AppSecret))
	if err != nil {
		return err
	}
	eerr := json.Unmarshal(b, &c.token)
	if eerr == nil && c.token.AccessToken != "" {
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

func (c *Client) Request(url string) ([]byte, *ClientError) {
	client := &http.Client{
		Timeout: c.config.DefaultTimeout,
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
