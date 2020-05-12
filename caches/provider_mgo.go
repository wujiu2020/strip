package caches

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/teapots/utils"
)

type MgoConfig struct {
	AutoExpire     bool
	KeyField       string
	ValueField     string
	ExpiredAtField string
	Connect        func(f func(c *mgo.Collection) error) error
}

type mgoProvide struct {
	config  *MgoConfig
	connect func(func(c *mgo.Collection) error) error
}

var _ CacheProvider = new(mgoProvide)

func NewMgoProvider(config MgoConfig) (prov CacheProvider, err error) {
	provider := new(mgoProvide)
	provider.config = &config
	provider.connect = config.Connect

	err = provider.connect(func(c *mgo.Collection) (err error) {
		err = c.Database.Session.Ping()
		if err != nil {
			err = fmt.Errorf("NewmgoProvide, Ping: %v", err)
			return
		}

		// unique index of key field
		if config.KeyField != "_id" {
			if err = c.EnsureIndex(mgo.Index{Key: []string{config.KeyField}, Name: config.KeyField, Unique: true}); err != nil {
				err = fmt.Errorf("NewmgoProvide, EnsureIndex: %v", err)
				return
			}
		}

		expireIndex := mgo.Index{
			Name: config.ExpiredAtField,
			Key:  []string{config.ExpiredAtField},
		}

		if config.AutoExpire {
			// 创建 mgo 自动过期的索引
			expireIndex.ExpireAfter = time.Minute
		}

		// index of expiredAt field
		if err = c.EnsureIndex(expireIndex); err != nil {
			err = fmt.Errorf("NewmgoProvide, EnsureIndex: %v", err)
			return
		}

		return
	})
	if err != nil {
		return
	}
	prov = provider
	return
}

func (p *mgoProvide) Get(key string) (value utils.StrTo, err error) {
	values, err := p.fetchValues(key)
	if err != nil {
		err = p.handleError(err)
		return
	}

	v, ok := p.getValue(values)
	if !ok {
		err = ErrMissedKey
		return
	}
	value = utils.StrTo(utils.ToStr(v))
	return
}

func (p *mgoProvide) Set(key string, val interface{}, params ...int) (err error) {
	err = p.connect(func(c *mgo.Collection) error {
		timeout := getTimeoutDurEx(params...)
		expiredAt := time.Now().Add(timeout)

		_, er := c.Upsert(bson.M{
			p.config.KeyField: key,
		}, bson.M{
			p.config.KeyField:       key,
			p.config.ValueField:     val,
			p.config.ExpiredAtField: expiredAt,
		})
		return er
	})
	return
}

func (p *mgoProvide) Touch(key string, params ...int) (err error) {
	err = p.connect(func(c *mgo.Collection) error {
		timeout := getTimeoutDurEx(params...)
		expiredAt := time.Now().Add(timeout)

		er := c.Update(bson.M{
			p.config.KeyField: key,
		}, bson.M{
			"$set": bson.M{
				p.config.ExpiredAtField: expiredAt,
			},
		})
		return er
	})
	err = p.handleError(err)
	return
}

func (p *mgoProvide) Delete(key string) (err error) {
	err = p.connect(func(c *mgo.Collection) error {
		er := c.Remove(bson.M{
			p.config.KeyField: key,
		})
		return er
	})
	err = p.handleError(err)
	return
}

func (p *mgoProvide) Incr(key string, params ...int) (err error) {
	cnt := 1
	if len(params) > 0 {
		cnt = params[0]
	}

	err = p.connect(func(c *mgo.Collection) error {
		er := c.Update(bson.M{
			p.config.KeyField: key,
		}, bson.M{
			"$inc": bson.M{
				p.config.ValueField: cnt,
			},
		})
		return er
	})
	err = p.handleError(err)
	return
}

func (p *mgoProvide) Decr(key string, params ...int) (err error) {
	cnt := -1
	if len(params) > 0 {
		cnt = 0 - params[0]
	}

	err = p.connect(func(c *mgo.Collection) error {
		er := c.Update(bson.M{
			p.config.KeyField: key,
		}, bson.M{
			"$inc": bson.M{
				p.config.ValueField: cnt,
			},
		})
		return er
	})
	err = p.handleError(err)
	return
}

func (p *mgoProvide) Has(key string) (bool, error) {
	values, err := p.fetchValues(key)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = nil
		}
		return false, err
	}

	_, ok := p.getValue(values)
	return ok, nil
}

func (p *mgoProvide) Clean() (err error) {
	err = p.connect(func(c *mgo.Collection) error {
		_, er := c.RemoveAll(nil)
		return er
	})
	return
}

func (p *mgoProvide) GC() (err error) {
	if p.config.AutoExpire {
		return
	}
	err = p.connect(func(c *mgo.Collection) error {
		_, er := c.RemoveAll(bson.M{
			p.config.ExpiredAtField: bson.M{
				"$lte": time.Now(),
			},
		})
		return er
	})
	return
}

func (p *mgoProvide) fetchValues(key string) (values map[string]interface{}, err error) {
	err = p.connect(func(c *mgo.Collection) error {
		er := c.Find(bson.M{
			p.config.KeyField: key,
		}).One(&values)
		return er
	})
	return values, err
}

func (p *mgoProvide) getValue(values map[string]interface{}) (value interface{}, ok bool) {
	if values == nil {
		return
	}

	// get expired time
	expiredAt, exists := values[p.config.ExpiredAtField].(time.Time)
	if exists {
		// not expired yet
		if time.Now().Before(expiredAt) {

			// get cached value
			value, ok = values[p.config.ValueField]
		}
	}
	return
}

func (p *mgoProvide) handleError(err error) error {
	if err == mgo.ErrNotFound {
		return ErrMissedKey
	}
	return err
}
