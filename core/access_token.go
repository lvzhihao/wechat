package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var API map[string]string = map[string]string{
	"TOKEN":      "https://api.weixin.qq.com/cgi-bin/token?grant_type=%s&appid=%s&secret=%s",
	"CALLBACKIP": "https://api.weixin.qq.com/cgi-bin/getcallbackip?access_token=%s",
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type Error struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type CallbackIpList struct {
	IpList []string `json:"ip_list"`
}

func (this *AccessToken) Token() string {
	return this.AccessToken
}

func (this *Error) Error() string {
	return fmt.Sprintf("ErrCode: %s; ErrMsg: %s", this.ErrCode, this.ErrMsg)
}

func init() {
	//todo
}

func GetAccessToken(appid, secret string) (*AccessToken, *Error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fmt.Sprintf(API["TOKEN"], "client_credential", appid, secret))
	if err != nil {
		return nil, &Error{ErrCode: -2, ErrMsg: err.Error()}
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &Error{ErrCode: -2, ErrMsg: err.Error()}
	}
	var ret AccessToken
	err = json.Unmarshal(b, &ret)
	if err == nil && ret.Token() != "" {
		return &ret, nil
	} else {
		var retErr Error
		err = json.Unmarshal(b, &retErr)
		if err == nil {
			return nil, &retErr
		} else {
			return nil, &Error{ErrCode: -2, ErrMsg: err.Error()}
		}
	}
}

func GetCallbackIpList(token string) (*CallbackIpList, *Error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fmt.Sprintf(API["CALLBACKIP"], token))
	if err != nil {
		return nil, &Error{ErrCode: -2, ErrMsg: err.Error()}
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &Error{ErrCode: -2, ErrMsg: err.Error()}
	}
	var ret CallbackIpList
	err = json.Unmarshal(b, &ret)
	if err == nil && len(ret.IpList) > 0 {
		return &ret, nil
	} else {
		var retErr Error
		err = json.Unmarshal(b, &retErr)
		if err == nil {
			return nil, &retErr
		} else {
			return nil, &Error{ErrCode: -2, ErrMsg: err.Error()}
		}
	}
}
