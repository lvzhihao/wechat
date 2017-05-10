package core_test

import (
	"flag"
	"log"
	"testing"
	"time"

	"github.com/lvzhihao/wechat/core"
)

var appId string
var appSecret string
var testClient *core.Client

func init() {
	flag.StringVar(&appId, "app_id", "", "第三方用户唯一凭证")
	flag.StringVar(&appSecret, "app_secret", "", "第三方用户唯一凭证密钥")
	flag.Parse()
	err := &core.ClientError{}
	testClient, err = core.New(&core.ClientConfig{
		AppId:          appId,
		AppSecret:      appSecret,
		DefaultTimeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func Test001NewClient(t *testing.T) {
	t.Log(testClient.FetchToken())
}
