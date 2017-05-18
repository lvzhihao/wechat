package models

import (
	"time"

	"github.com/lvzhihao/wechat/core"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func UserAccessTokenEnsureIndex(s *mgo.Session) {
	c := s.DB("").C("user_access_token")
	c.EnsureIndex(mgo.Index{
		Key:        []string{"appid", "openid"},
		Unique:     true,
		DropDups:   true,
		Background: true, // See notes.
		Sparse:     true,
	})
}

type UserAccessToken struct {
	AppId       string               `bson:"appid" json:"-"`
	OpenId      string               `bson:"openid" json:"openid"`
	AccessToken core.UserAccessToken `bson:"access_token" json:"access_token"`
	UpdatedTime time.Time            `bson:"updated_time" json:"-"`
}

func (q *UserAccessToken) Upsert(s *mgo.Session) error {
	c := s.DB("").C("user_access_token")
	q.UpdatedTime = time.Now() //updated time
	_, err := c.Upsert(bson.M{"appid": q.AppId, "openid": q.OpenId}, q)
	return err
}
