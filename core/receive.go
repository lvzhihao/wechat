package core

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type MsgComm struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   int    `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	MsgId        int64  `xml:"MsgId"`
}

type Msg struct {
	MsgComm
	Content      string `xml:"Content"`
	PicUrl       string `xml:"PicUrl"`
	MediaId      string `xml:"MediaId"`
	Format       string `xml:"Format"`
	ThumbMediaId string `xml:"ThumbMediaId"`
	LocationX    string `xml:"Location_X"`
	LocationY    string `xml:"Location_Y"`
	Scale        string `xml:"Scale"`
	Label        string `xml:"Label"`
	Title        string `xml:"Title"`
	Description  string `xml:"Description"`
	Url          string `xml:"Url"`
	Event        string `xml:"Event"`
	EventKey     string `xml:"EventKey"`
	Ticket       string `xml:"Ticket"`
	Latitude     string `xml:"Latitude"`
	Longitude    string `xml:"Longitude"`
	Precision    string `xml:"Precision"`
}

type RetMsgComm struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   int    `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
}

type RetTextMsg struct {
	XMLName xml.Name `xml:"xml"`
	RetMsgComm
	Content string `xml:"Content"`
}

type RetImageMsg struct {
	XMLName xml.Name `xml:"xml"`
	RetMsgComm
	MediaId string `xml:"Image>MediaId"`
}

type RetVoiceMsg struct {
	XMLName xml.Name `xml:"xml"`
	RetMsgComm
	MediaId string `xml:"Voice>MediaId"`
}

type RetVideoMsg struct {
	XMLName xml.Name `xml:"xml"`
	RetMsgComm
	MediaId     string `xml:"Video>MediaId"`
	Title       string `xml:"Video>Title"`
	Description string `xml:"Video>Description"`
}

type RetMusicMsg struct {
	XMLName xml.Name `xml:"xml"`
	RetMsgComm
	Title        string `xml:"Music>Title"`
	Description  string `xml:"Music>Description"`
	MusicUrl     string `xml:"Music>MusicUrl"`
	HQMusicUrl   string `xml:"Music>HQMusicUrl"`
	ThumbMediaId string `xml:"Music>ThumbMediaId"`
}

type RetNewsMsg struct {
	XMLName xml.Name `xml:"xml"`
	RetMsgComm
	ArticleCount int           `xml:"ArticleCount"`
	Item         []RetNewsItem `xml:"Articles>item"`
}

type RetNewsItem struct {
	Title       string `xml:"Title"`
	Description string `xml:"Description"`
	PicUrl      string `xml:"PicUrl"`
	Url         string `xml:"Url"`
}

func ReceiveMsgCheckSign(token string, r *http.Request) bool {
	q := r.URL.Query()
	var params []string
	params = []string{"eyNhcbVVQvFyyAm7ojZvmuuaxcTjRhEWqRErwi6AQKpEuYdfnM", q.Get("timestamp"), q.Get("nonce")}
	sort.Sort(sort.StringSlice(params))
	signStr := strings.Join(params, "")
	sign := fmt.Sprintf("%x", sha1.Sum([]byte(signStr)))
	if strings.Compare(q.Get("signature"), sign) == 0 {
		return true
	} else {
		return false
	}
}
