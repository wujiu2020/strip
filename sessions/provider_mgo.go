package sessions

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/teapots/caches"
)

type MgoConfig struct {
	Config
	AutoExpire bool // is support auto expire?
	Connect    func(func(c *mgo.Collection) error) error
}

func NewMgoProvider(config MgoConfig) (prov SessionProvider, err error) {
	cache, err := caches.NewMgoProvider(caches.MgoConfig{
		KeyField:       keyField,
		ValueField:     valueField,
		ExpiredAtField: expiredAtField,
		AutoExpire:     config.AutoExpire,
		Connect:        config.Connect,
	})
	if err != nil {
		return
	}
	prov = newSessProvide(config.Config, cache, &bsonConverter{})
	return
}

type bsonConverter struct{}

func (p *bsonConverter) Marshal(in interface{}) (out []byte, err error) {
	return bson.Marshal(in)
}

func (p *bsonConverter) Unmarshal(in []byte, out interface{}) (err error) {
	return bson.Unmarshal(in, out)
}
