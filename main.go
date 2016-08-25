package main

import (
	"log"
	"os"

	"github.com/lvzhihao/wechat/core"
)

var appid string = "wx09f09e8e0ea397ca"
var appsecret string = "81ac48e5d9682df50b6d82a0b62f27b6"
var token string = ""

func init() {

}

func main() {
	t, err := core.GetAccessToken(appid, appsecret)
	//fmt.Println(t, err)
	if err != nil {
		log.Println("[ERR]", err)
		os.Exit(2)
	} else {
		log.Println("[DEBUG]", t.Token())
		token = t.Token()
	}
	ipList, err := core.GetCallbackIpList(token)
	if err != nil {
		log.Println("[ERR]", err)
		os.Exit(2)
	} else {
		log.Println("[DEBUG]", ipList.IpList)
	}
}
