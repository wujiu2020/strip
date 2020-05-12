package sessions

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/teapots/caches"
)

type McConfig struct {
	Config
	KeyPrefix string
	Client    *memcache.Client
}

func NewMcProvider(config McConfig) (prov SessionProvider, err error) {
	cache, err := caches.NewMcProvider(caches.McConfig{
		KeyPrefix: config.KeyPrefix,
		Client:    config.Client,
	})
	if err != nil {
		return
	}
	prov = newSessProvide(config.Config, cache, &bsonConverter{})
	return
}
