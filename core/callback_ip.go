package core

import (
	"encoding/json"
	"fmt"
)

type CallbackIpList struct {
	IpList []string `json:"ip_list"`
}

func (c *Client) GetCallbackIpList() (*CallbackIpList, *ClientError) {
	b, err := c.Request(fmt.Sprintf(BaseApis["CALLBACKIP"], c.FetchToken()))
	if err != nil {
		return nil, err
	}
	var ret CallbackIpList
	eerr := json.Unmarshal(b, &ret)
	if eerr == nil && len(ret.IpList) > 0 {
		return &ret, nil
	} else {
		var retErr ClientError
		eerr = json.Unmarshal(b, &retErr)
		if err == nil {
			return nil, &retErr
		} else {
			return nil, &ClientError{ErrCode: -2, ErrMsg: eerr.Error()}
		}
	}
}
