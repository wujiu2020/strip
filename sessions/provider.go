package sessions

import (
	"github.com/teapots/caches"
	"github.com/teapots/utils"
)

const (
	keyField       = "_id"
	valueField     = "value"
	expiredAtField = "expiredAt"
)

type converter interface {
	Marshal(in interface{}) (out []byte, err error)
	Unmarshal(in []byte, out interface{}) (err error)
}

type sessStore struct {
	Sid    string                 `bson:"_id"`
	Values map[string]interface{} `bson:"value"`
}

func newSessStore(sid string, params ...map[string]interface{}) *sessStore {
	store := &sessStore{
		Sid: sid,
	}

	if len(params) > 0 {
		store.Values = params[0]
	}

	if store.Values == nil {
		store.Values = make(map[string]interface{})
	}
	return store
}

type sessData struct {
	store    *sessStore
	provider *sessProvide

	changed bool // flag when values has changed
	closed  bool // flag when data destroy
}

var _ SessionStore = new(sessData)

func newSessData(p *sessProvide, store *sessStore, changed bool) *sessData {
	sess := &sessData{
		store:    store,
		provider: p,
		changed:  changed,
	}

	if store.Values == nil {
		store.Values = make(map[string]interface{})
	}
	return sess
}

// get session id
func (m *sessData) Sid() string {
	return m.store.Sid
}

func (m *sessData) Set(key string, value interface{}) {
	m.changed = true
	m.store.Values[key] = value
}

func (m *sessData) Get(key string) *utils.Value {
	if v, ok := m.store.Values[key]; ok {
		return utils.ValueTo(v)
	}
	return utils.ValueTo(nil)
}

func (m *sessData) Delete(key string) {
	m.changed = true
	delete(m.store.Values, key)
}

func (m *sessData) Has(key string) bool {
	_, ok := m.store.Values[key]
	return ok
}

// duplicate all values
func (m *sessData) Values() map[string]interface{} {
	values := make(map[string]interface{}, len(m.store.Values))
	for key, value := range m.store.Values {
		values[key] = value
	}
	return values
}

// clear all values in session
func (m *sessData) Clean() {
	m.changed = true
	m.store.Values = make(map[string]interface{})
}

// save session values to store
func (m *sessData) Flush() error {
	// has destroy
	if m.closed {
		return nil
	}

	// no changes
	if !m.changed {
		return nil
	}

	m.changed = false
	return m.provider.save(m.store)
}

// destory session values in store
func (m *sessData) Destroy() error {
	m.closed = true
	m.store.Values = make(map[string]interface{})
	return m.provider.Destroy(m.store.Sid)
}

func (m *sessData) Touch() error {
	return m.provider.cache.Touch(m.store.Sid, m.provider.config.SessionExpire)
}

type sessProvide struct {
	config Config
	cache  caches.CacheProvider
	conv   converter
}

var _ SessionProvider = new(sessProvide)

func newSessProvide(config Config, cache caches.CacheProvider, conv converter) *sessProvide {
	return &sessProvide{
		config: config,
		cache:  cache,
		conv:   conv,
	}
}

func (p *sessProvide) Create(sid string, params ...map[string]interface{}) (sess SessionStore, err error) {
	sess = p.create(sid, params...)
	return
}

func (p *sessProvide) Read(sid string) (sess SessionStore, err error) {
	value, err := p.cache.Get(sid)
	if err != nil {
		if err == caches.ErrMissedKey {
			err = ErrNotFoundSession
		}
		return
	}

	var values map[string]interface{}
	err = p.conv.Unmarshal(value.Bytes(), &values)
	if err != nil {
		return
	}
	store := &sessStore{Sid: sid, Values: values}
	sess = newSessData(p, store, false)
	return
}

// 为现有的session数据更换sid
func (p *sessProvide) Regenerate(old string, sid string) (sess SessionStore, err error) {
	value, err := p.cache.Get(old)
	if err != nil {
		if err == caches.ErrMissedKey {
			err = ErrNotFoundSession
		}
		return
	}

	err = p.cache.Set(sid, value, p.config.SessionExpire)
	if err != nil {
		return
	}

	p.Destroy(old)
	return p.Read(sid)
}

func (p *sessProvide) Destroy(sid string) (err error) {
	err = p.cache.Delete(sid)
	return
}

func (p *sessProvide) GC() (err error) {
	err = p.cache.GC()
	return
}

func (p *sessProvide) Config() *Config {
	cfg := p.config
	return &cfg
}

func (p *sessProvide) create(sid string, params ...map[string]interface{}) (sess SessionStore) {
	store := newSessStore(sid, params...)
	sess = newSessData(p, store, len(params) > 0)
	return
}

func (p *sessProvide) save(store *sessStore) (err error) {
	value, err := p.conv.Marshal(store.Values)
	if err != nil {
		return
	}
	err = p.cache.Set(store.Sid, value, p.config.SessionExpire)
	return
}
