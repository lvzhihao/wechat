package core

import (
	"flag"
	"log"
	"testing"
	"time"
)

var appId string
var appSecret string
var testClient *Client

func init() {
	flag.StringVar(&appId, "app_id", "", "第三方用户唯一凭证")
	flag.StringVar(&appSecret, "app_secret", "", "第三方用户唯一凭证密钥")

}

func Test_001_NewClient(t *testing.T) {
	err := &ClientError{}
	testClient, err = New(&ClientConfig{
		AppId:          appId,
		AppSecret:      appSecret,
		DefaultTimeout: 10 * time.Second,
	})
	if err != nil {
		log.Panic(err)
	}
	t.Log(testClient.FetchToken())
}
