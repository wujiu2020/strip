package caches

import (
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/teapots/utils"
)

type RedisConfig struct {
	KeyPrefix string
	Client    *pool.Pool
}

type redisProvide struct {
	RedisConfig
}

var _ CacheProvider = new(redisProvide)

func NewRedisProvider(config RedisConfig) (sess CacheProvider, err error) {
	provider := new(redisProvide)
	provider.RedisConfig = config
	return provider, nil
}

func (p *redisProvide) Get(key string) (value utils.StrTo, err error) {
	key = p.KeyPrefix + key

	conn, err := p.Client.Get()
	if err != nil {
		return
	}
	defer p.Client.Put(conn)

	resp := conn.Cmd("GET", key)
	v, err := resp.Str()
	if err != nil {
		err = p.handleError(err)
		return
	}
	value = utils.StrTo(v)
	return
}

func (p *redisProvide) Set(key string, val interface{}, params ...int) (err error) {
	key = p.KeyPrefix + key
	timeout := getTimeoutDur(params...)

	conn, err := p.Client.Get()
	if err != nil {
		return
	}
	defer p.Client.Put(conn)

	args := []interface{}{key, utils.ToStr(val)}
	if timeout > 0 {
		args = append(args, "EX", timeout.Seconds())
	}
	resp := conn.Cmd("SET", args...)
	err = resp.Err
	return
}

func (p *redisProvide) Touch(key string, params ...int) (err error) {
	key = p.KeyPrefix + key
	timeout := getTimeoutDur(params...)

	conn, err := p.Client.Get()
	if err != nil {
		return
	}
	defer p.Client.Put(conn)

	resp := conn.Cmd("EXPIRE", key, timeout.Seconds())
	touched, err := resp.Int()
	if err != nil {
		err = p.handleError(err)
		return
	}
	if touched == 0 {
		err = ErrMissedKey
	}
	return
}

func (p *redisProvide) Delete(key string) (err error) {
	key = p.KeyPrefix + key

	conn, err := p.Client.Get()
	if err != nil {
		return
	}
	defer p.Client.Put(conn)

	resp := conn.Cmd("DEL", key)
	deleted, err := resp.Int()
	if err != nil {
		err = p.handleError(err)
		return
	}
	if deleted == 0 {
		err = ErrMissedKey
	}
	return
}

func (p *redisProvide) Incr(key string, params ...int) (err error) {
	key = p.KeyPrefix + key
	cnt := 1
	if len(params) > 0 {
		cnt = params[0]
	}

	conn, err := p.Client.Get()
	if err != nil {
		return
	}
	defer p.Client.Put(conn)

	resp := conn.Cmd("INCRBY", key, cnt)
	err = resp.Err
	err = p.handleError(err)
	return
}

func (p *redisProvide) Decr(key string, params ...int) (err error) {
	key = p.KeyPrefix + key
	cnt := 1
	if len(params) > 0 {
		cnt = params[0]
	}

	conn, err := p.Client.Get()
	if err != nil {
		return
	}
	defer p.Client.Put(conn)

	resp := conn.Cmd("DECRBY", key, cnt)
	err = resp.Err
	err = p.handleError(err)
	return
}

func (p *redisProvide) Has(key string) (exists bool, err error) {
	key = p.KeyPrefix + key

	conn, err := p.Client.Get()
	if err != nil {
		return
	}
	defer p.Client.Put(conn)

	resp := conn.Cmd("EXISTS", key)
	e, err := resp.Int()
	if err == redis.ErrRespNil {
		err = nil
	}
	exists = e == 1
	return
}

func (p *redisProvide) Clean() (err error) {
	conn, err := p.Client.Get()
	if err != nil {
		return
	}
	defer p.Client.Put(conn)

	resp := conn.Cmd("FLUSHDB")
	err = resp.Err
	return
}

func (p *redisProvide) GC() error {
	return nil
}

func (p *redisProvide) handleError(err error) error {
	if err == redis.ErrRespNil {
		return ErrMissedKey
	}
	return err
}
