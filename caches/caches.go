package caches

import (
	"errors"
	"time"

	"github.com/wujiu2020/strip/utils"
)

const (
	DefaultTimeout = time.Minute
)

var (
	ErrMissedKey = errors.New("err_missed_key")
)

type CacheProvider interface {
	Get(key string) (utils.StrTo, error)                    // get cached value by key
	Set(key string, value interface{}, params ...int) error // set cached key, value with optional timeout seconds
	Touch(key string, params ...int) error                  // touch cached with new timeout seconds
	Delete(key string) error                                // delete cached value by key
	Incr(key string, params ...int) error                   // incr integer value
	Decr(key string, params ...int) error                   // decr integer value
	Has(key string) (bool, error)                           // check cached key exists
	Clean() error                                           // clean all cached values
	GC() error                                              // use for interval GC
}

func getTimeoutDur(params ...int) time.Duration {
	var timeout time.Duration
	if len(params) > 0 {
		switch {
		case params[0] > 0:
			timeout = time.Duration(params[0]) * time.Second
			return timeout
		default:
			timeout = time.Duration(params[0])
		}
	}

	if timeout == 0 {
		timeout = DefaultTimeout
	}
	if timeout < 0 {
		timeout = 0
	}
	return timeout
}

func getTimeoutDurEx(params ...int) time.Duration {
	timeout := getTimeoutDur(params...)
	if timeout == 0 {
		timeout = time.Hour * 24 * 9999
	}
	return timeout
}
