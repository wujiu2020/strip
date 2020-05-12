package sessions

import (
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/wujiu2020/strip/caches"
)

type RedisConfig struct {
	Config
	KeyPrefix string
	Client    *pool.Pool
}

func NewRedisProvider(config RedisConfig) (prov SessionProvider, err error) {
	cache, err := caches.NewRedisProvider(caches.RedisConfig{
		KeyPrefix: config.KeyPrefix,
		Client:    config.Client,
	})
	if err != nil {
		return
	}
	prov = newSessProvide(config.Config, cache, &bsonConverter{})
	return
}
