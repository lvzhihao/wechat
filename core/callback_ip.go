package core

import (
	"encoding/json"
	"fmt"
)

type CallbackIpList struct {
	IpList []string `json:"ip_list"`
}

func (c *Client) GetCallbackIpList() (*CallbackIpList, *ClientError) {
	b, eerr := c.Request(fmt.Sprintf(BaseApis["CALLBACKIP"], c.FetchToken()))
	if eerr != nil {
		return nil, eerr
	}
	var ret CallbackIpList
	err := json.Unmarshal(b, &ret)
	if err == nil && len(ret.IpList) > 0 {
		return &ret, nil
	} else {
		return nil, &ClientError{ErrCode: -2, ErrMsg: err.Error()}
	}
}
