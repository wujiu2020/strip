package caches

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/teapots/utils"
)

type McConfig struct {
	KeyPrefix string
	Client    *memcache.Client
}

type mcProvide struct {
	McConfig
}

var _ CacheProvider = new(mcProvide)

func NewMcProvider(config McConfig) (sess CacheProvider, err error) {
	provider := new(mcProvide)
	provider.McConfig = config
	return provider, nil
}

func (p *mcProvide) Get(key string) (value utils.StrTo, err error) {
	key = p.KeyPrefix + key
	item, err := p.Client.Get(key)
	if err != nil {
		err = p.handleError(err)
		return
	}
	value = utils.StrTo(string(item.Value))
	return
}

func (p *mcProvide) Set(key string, val interface{}, params ...int) (err error) {
	key = p.KeyPrefix + key
	timeout := getTimeoutDur(params...)
	err = p.Client.Set(&memcache.Item{
		Key:        key,
		Value:      []byte(utils.ToStr(val)),
		Expiration: int32(timeout.Seconds()),
	})
	return
}

func (p *mcProvide) Touch(key string, params ...int) (err error) {
	key = p.KeyPrefix + key
	timeout := getTimeoutDur(params...)
	err = p.Client.Touch(key, int32(timeout.Seconds()))
	err = p.handleError(err)
	return
}

func (p *mcProvide) Delete(key string) (err error) {
	key = p.KeyPrefix + key
	err = p.Client.Delete(key)
	err = p.handleError(err)
	return
}

func (p *mcProvide) Incr(key string, params ...int) (err error) {
	key = p.KeyPrefix + key
	cnt := 1
	if len(params) > 0 {
		cnt = params[0]
	}
	_, err = p.Client.Increment(key, uint64(cnt))
	err = p.handleError(err)
	return
}

func (p *mcProvide) Decr(key string, params ...int) (err error) {
	key = p.KeyPrefix + key
	cnt := 1
	if len(params) > 0 {
		cnt = params[0]
	}
	_, err = p.Client.Decrement(key, uint64(cnt))
	err = p.handleError(err)
	return
}

func (p *mcProvide) Has(key string) (bool, error) {
	key = p.KeyPrefix + key
	_, err := p.Client.Get(key)
	exists := err == nil
	if err == memcache.ErrCacheMiss {
		err = nil
	}
	return exists, err
}

func (p *mcProvide) Clean() error {
	return p.Client.FlushAll()
}

func (p *mcProvide) GC() error {
	return nil
}

func (p *mcProvide) handleError(err error) error {
	if err == memcache.ErrCacheMiss {
		return ErrMissedKey
	}
	return err
}
