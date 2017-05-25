package models

import (
	"time"

	"github.com/lvzhihao/wechat/core"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func UserInfoEnsureIndex(s *mgo.Session) {
	c := s.DB("").C(UserInfo{}.TableName())
	c.EnsureIndex(mgo.Index{
		Key:        []string{"appid", "openid"},
		Unique:     true,
		DropDups:   true,
		Background: true, // See notes.
		Sparse:     true,
	})
}

type UserInfo struct {
	AppId       string        `bson:"appid" json:"-"`
	OpenId      string        `bson:"openid" json:"openid"`
	Info        core.UserInfo `bson:"info" json:"info"`
	UpdatedTime time.Time     `bson:"updated_time" json:"-"`
}

func (q UserInfo) TableName() string {
	return "user_info"
}

func (q *UserInfo) Upsert(s *mgo.Session) error {
	c := s.DB("").C(UserInfo{}.TableName())
	q.UpdatedTime = time.Now() //updated time
	_, err := c.Upsert(bson.M{"appid": q.AppId, "openid": q.OpenId}, q)
	return err
}
