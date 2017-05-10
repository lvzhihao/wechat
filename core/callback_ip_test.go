package core_test

import (
	"testing"

	"github.com/lvzhihao/wechat/core"
)

func Test002CallbackIp(t *testing.T) {
	ips, err := core.GetCallbackIpList(testClient)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(ips)
	}
}
