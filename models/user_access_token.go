package models

import (
	"time"

	"github.com/lvzhihao/wechat/core"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func UserAccessTokenEnsureIndex(s *mgo.Session) {
	c := s.DB("").C(UserAccessToken{}.TableName())
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

func (q UserAccessToken) TableName() string {
	return "user_access_token"
}

func (q *UserAccessToken) Upsert(s *mgo.Session) error {
	c := s.DB("").C(UserAccessToken{}.TableName())
	q.UpdatedTime = time.Now() //updated time
	_, err := c.Upsert(bson.M{"appid": q.AppId, "openid": q.OpenId}, q)
	return err
}

func ListUserAccessToken(s *mgo.Session) (list []*UserAccessToken, err error) {
	c := s.DB("").C(UserAccessToken{}.TableName())
	err = c.Find(bson.M{
		"updated_time": bson.M{"$lte": time.Now().Add(-110 * time.Minute)},
	}).Sort("updated_time").All(&list)
	return
}
