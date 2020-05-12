package caches

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/stretchr/testify/assert"
)

var (
	cache CacheProvider
)

func initMgoProvider() {
	conn, err := mgo.Dial("mongodb://localhost")
	if err != nil {
		log.Fatal(err)
	}

	cache, err = NewMgoProvider(MgoConfig{
		KeyField:       "_id",
		ValueField:     "value",
		ExpiredAtField: "expiredAt",
		AutoExpire:     true,
		Connect: func(f func(c *mgo.Collection) error) error {
			return f(conn.DB("test_caches").C("caches"))
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func initMcProvider() {
	client := memcache.New("localhost:11211")

	var err error
	cache, err = NewMcProvider(McConfig{
		KeyPrefix: "prefix",
		Client:    client,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func initRedisProvider() {
	client, err := pool.New("tcp", "localhost:6379", 10)
	if err != nil {
		log.Fatal(err)
	}

	cache, err = NewRedisProvider(RedisConfig{
		KeyPrefix: "prefix",
		Client:    client,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	initMgoProvider()
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	initMcProvider()
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	initRedisProvider()
	os.Exit(m.Run())
}

func Test_CacheProvider(t *testing.T) {
	var err error

	err = cache.Set("key", "val")
	assert.NoError(t, err)

	has, err := cache.Has("key")
	assert.NoError(t, err)
	assert.True(t, has)

	value, err := cache.Get("key")
	assert.NoError(t, err)
	assert.Equal(t, value.String(), "val")

	value, err = cache.Get("undefined")
	assert.Equal(t, ErrMissedKey, err)

	err = cache.Touch("undefined", 10)
	assert.Equal(t, ErrMissedKey, err)

	err = cache.Delete("undefined")
	assert.Equal(t, ErrMissedKey, err)

	has, err = cache.Has("undefined")
	assert.NoError(t, err)
	assert.False(t, has)

	cache.Set("key", "none")

	has, err = cache.Has("key")
	assert.NoError(t, err)
	assert.True(t, has)

	value, err = cache.Get("key")
	assert.NoError(t, err)
	assert.Equal(t, value.String(), "none")

	err = cache.Touch("key")
	assert.NoError(t, err)

	cache.Delete("key")
	has, err = cache.Has("key")
	assert.NoError(t, err)
	assert.False(t, has)

	value, err = cache.Get("key")
	assert.Equal(t, ErrMissedKey, err)

	err = cache.Set("key1", "val")
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	err = cache.Set("key2", "val")
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	err = cache.Set("key3", "val")
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	has, _ = cache.Has("key1")
	assert.True(t, has)
	has, _ = cache.Has("key2")
	assert.True(t, has)
	has, _ = cache.Has("key3")
	assert.True(t, has)
	cache.Clean()
	has, _ = cache.Has("key1")
	assert.False(t, has)
	has, _ = cache.Has("key2")
	assert.False(t, has)
	has, _ = cache.Has("key3")
	assert.False(t, has)

	cache.Set("key", "val", 60)
	has, _ = cache.Has("key")
	assert.True(t, has)

	cache.Set("key", "val", -1)
	has, _ = cache.Has("key")
	assert.True(t, has)

	err = cache.Set("key", "val", 0)
	assert.NoError(t, err)
	has, _ = cache.Has("key")
	assert.True(t, has)

	cache.Set("num", 1)
	value, err = cache.Get("num")
	assert.NoError(t, err)
	assert.Equal(t, 1, value.MustInt())

	err = cache.Incr("num", 1)
	assert.NoError(t, err)

	value, err = cache.Get("num")
	assert.NoError(t, err)
	assert.Equal(t, 2, value.MustInt())

	err = cache.Decr("num", 1)
	assert.NoError(t, err)

	value, err = cache.Get("num")
	assert.NoError(t, err)
	assert.Equal(t, 1, value.MustInt())
}

func Test_CacheMgoKeyExpired(t *testing.T) {
	mc, ok := cache.(*mgoProvide)
	if !ok {
		t.Skip()
	}

	err := mc.connect(func(c *mgo.Collection) error {
		return c.Update(bson.M{
			mc.config.KeyField: "key",
		}, bson.M{
			"$set": bson.M{
				mc.config.ExpiredAtField: time.Now(),
			},
		})
	})
	assert.NoError(t, err)

	has, _ := cache.Has("key")
	assert.False(t, has)
}
