package core_test

import (
	"testing"
)

func Test002CallbackIp(t *testing.T) {
	ips, err := testClient.GetCallbackIpList()
	if err != nil {
		t.Error(err)
	} else {
		t.Log(ips)
	}
}
