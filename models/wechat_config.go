package models

import (
	"errors"
	"time"

	"github.com/lvzhihao/wechat/utils"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type WechatConfig struct {
	Name         string             `bson:"name" json:"name"`
	AppId        string             `bson:"appid" json:"appid"`
	AppSecret    string             `bson:"appsecret" json:"appsecret"`
	ReceiveToken string             `bson:"receive_token" json:"receive_token"`
	CallbackUrl  string             `bson:"callback_url" json:"callback_url"`
	Detail       WechatConfigDetail `bson:"detail" json:"detail"`
	CreatedTime  time.Time          `bson:"created_time" json:"-"`
	UpdatedTime  time.Time          `bson:"updated_time" json:"-"`
	DeletedTime  time.Time          `bson:"deleted_time" json:"-"`
}

func WechatConfigEnusreConfig(s *mgo.Session) {
	c := s.DB("").C("wechat_config")
	c.EnsureIndex(mgo.Index{
		Key:        []string{"name"},
		Unique:     true,
		DropDups:   true,
		Background: true, // See notes.
		Sparse:     true,
	})
	c.EnsureIndex(mgo.Index{
		Key:        []string{"appid"},
		Unique:     true,
		DropDups:   true,
		Background: true, // See notes.
		Sparse:     true,
	})
}

type WechatConfigDetail struct {
	Key    string `bson:"key" json:"key"`
	Secret string `bson:"secret" json:"secret"`
	//todo acl
}

func (q *WechatConfigDetail) Ensure() {
	if q.Key == "" {
		q.Key = utils.RandStringRunes(8)
		q.Secret = utils.RandStringRunes(63)
	}
}

func (q *WechatConfig) DetailEnsure(s *mgo.Session) error {
	q.Detail.Ensure()
	return q.Save(s)
}

func (q *WechatConfig) Fetch(s *mgo.Session) error {
	c := s.DB("").C("wechat_config")
	err := c.Find(bson.M{"name": q.Name}).One(q)
	return err
}

func (q *WechatConfig) Save(s *mgo.Session) error {
	c := s.DB("").C("wechat_config")
	err := c.Update(bson.M{"name": q.Name}, q)
	return err
}

func (q *WechatConfig) Create(s *mgo.Session) error {
	c := s.DB("").C("wechat_config")
	if q.Name == "" {
		return errors.New("name empty")
	}
	q.CreatedTime = time.Now()
	return c.Insert(q)
}

func ListWechatConfig(s *mgo.Session) (list []*WechatConfig, err error) {
	c := s.DB("").C("wechat_config")
	err = c.Find(bson.M{}).Sort("created_time").All(&list)
	return
}
