package sessions

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/globalsign/mgo"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/stretchr/testify/assert"
)

var (
	manager *SessionManager

	cookieConfig = &CookieConfig{
		CookieName:         "DOMAIN_SESSION",
		CookieRememberName: "DOMAIN_REMEMBER",
		CookieDomain:       ".domain.com",
		CookieSecure:       true,
		CookieExpire:       3600,
		RememberExpire:     3600,
	}
	config = Config{
		SessionExpire: 3600,
		SecretKey:     "secret_key",
	}
)

func initMgoProvider() {
	conn, err := mgo.Dial("mongodb://localhost")
	if err != nil {
		log.Fatal(err)
	}

	provider, err := NewMgoProvider(MgoConfig{
		Config:     config,
		AutoExpire: true,
		Connect: func(f func(c *mgo.Collection) error) error {
			return f(conn.DB("test_sessions").C("sessions"))
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	manager = NewSessionManager(provider)
}

func initMemcacheProvider() {
	client := memcache.New("localhost:11211")
	provider, err := NewMcProvider(McConfig{
		Config:    config,
		KeyPrefix: "prefix",
		Client:    client,
	})
	if err != nil {
		log.Fatal(err)
	}
	manager = NewSessionManager(provider)
}

func initRedisProvider() {
	client, err := pool.New("tcp", "localhost:6379", 10)
	if err != nil {
		log.Fatal(err)
	}
	provider, err := NewRedisProvider(RedisConfig{
		Config:    config,
		KeyPrefix: "prefix",
		Client:    client,
	})
	if err != nil {
		log.Fatal(err)
	}
	manager = NewSessionManager(provider)
}

func TestMain(m *testing.M) {
	initMgoProvider()
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	initMemcacheProvider()
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	initRedisProvider()
	os.Exit(m.Run())
}

func Test_SessionManager(t *testing.T) {
	sid := CreateSid()
	r := createNewSessionRequest(sid, time.Now())
	w := httptest.NewRecorder()

	sess, _, err := manager.Start(cookieConfig, w, r)
	assert.NoError(t, err)
	assert.Equal(t, sid, sess.Sid())

	sid = sess.Sid()
	r = createNewSessionRequest(sid, time.Now())

	err = manager.Destroy(cookieConfig, w, r)
	assert.NoError(t, err)

	// Session 删除以后，重新生成
	sess, _, err = manager.Start(cookieConfig, w, r)
	assert.NoError(t, err)
	assert.Equal(t, sid, sess.Sid())

	_, err = manager.provider.Read(sid)
	assert.Equal(t, ErrNotFoundSession, err)
}

func Test_SessionStore(t *testing.T) {
	sid := CreateSid()
	r := createNewSessionRequest(sid, time.Now())
	w := httptest.NewRecorder()

	sess, _, err := manager.Start(cookieConfig, w, r)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	sid = sess.Sid()

	sess.Set("uid", 110)
	assert.Equal(t, sess.Get("uid").MustInt(), 110)

	sess.Clean()
	assert.Equal(t, sess.Get("uid").Value(), nil)

	sess.Set("uid", 110)
	sess.Delete("uid")

	sess.Clean()
	assert.Equal(t, sess.Get("uid").Value(), nil)

	sess.Set("uid", 110)
	err = sess.Flush()
	assert.Equal(t, err, nil)

	r = createNewSessionRequest(sid, time.Now())
	sess, _, err = manager.Start(cookieConfig, w, r)
	assert.NoError(t, err)
	assert.Equal(t, sess.Get("uid").MustInt(), 110)

	sess, err = manager.provider.Read(sid)
	assert.NoError(t, err)
	assert.Equal(t, sess.Get("uid").MustInt(), 110)

	err = sess.Destroy()
	assert.Equal(t, err, nil)

	err = sess.Flush()
	assert.Equal(t, err, nil)

	_, err = manager.provider.Read(sid)
	assert.Equal(t, err == ErrNotFoundSession, true)
}

func Test_SessionRegenerate(t *testing.T) {
	sid := CreateSid()
	r := createNewSessionRequest(sid, time.Now())
	w := httptest.NewRecorder()

	sess, _, err := manager.Start(cookieConfig, w, r)
	assert.NoError(t, err)
	assert.Equal(t, sid, sess.Sid())

	sid = sess.Sid()
	sess, _, err = manager.Regenerate(cookieConfig, w, r, nil)
	assert.NoError(t, err)
	assert.NotEqual(t, sid, sess.Sid())

	sid = sess.Sid()
	r = createNewSessionRequest(sid, time.Now())

	sess.Set("key", "value")
	err = sess.Flush()
	assert.NoError(t, err)

	sess, err = manager.provider.Read(sid)
	assert.NoError(t, err)

	sess, err = manager.provider.Read(sid)
	assert.NoError(t, err)

	sess, _, err = manager.Regenerate(cookieConfig, w, r, nil)
	assert.NoError(t, err)

	nsid := sess.Sid()

	sess, err = manager.provider.Read(sid)
	assert.Equal(t, ErrNotFoundSession, err)

	sess, err = manager.provider.Read(nsid)
	assert.NoError(t, err)
	assert.Equal(t, nsid, sess.Sid())
	assert.Equal(t, "value", sess.Get("key").String())
}

func Test_SessionExpire(t *testing.T) {
	sid := CreateSid()
	r := createNewSessionRequest(sid, time.Now())
	w := httptest.NewRecorder()

	sess, _, err := manager.Start(cookieConfig, w, r)
	assert.NoError(t, err)

	sid = sess.Sid()

	sess, err = manager.provider.Read(sid)
	assert.Equal(t, ErrNotFoundSession, err)
	assert.Equal(t, nil, sess)
}

func Test_SessionCreatedAt(t *testing.T) {
	sid := CreateSid()
	r := createNewSessionRequest(sid, time.Now())
	w := httptest.NewRecorder()

	sess, createdAt, err := manager.Start(cookieConfig, w, r)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	r = createNewSessionRequest(sess.Sid(), createdAt)

	_, parsedTime, _ := manager.Start(cookieConfig, w, r)

	assert.Equal(t, createdAt.UnixNano(), parsedTime.UnixNano())
}

func createNewSessionRequest(sid string, createdAt time.Time) *http.Request {
	value, _ := EncodeSecureValue(sid, config.SecretKey, createdAt)

	cookie := &http.Cookie{
		Name:   cookieConfig.CookieName,
		Domain: cookieConfig.CookieDomain,
		Value:  value,
		Path:   "/",
		Secure: true,
		MaxAge: config.SessionExpire,
	}
	r, _ := http.NewRequest("GET", "https://www.domain.com", nil)
	r.AddCookie(cookie)
	return r
}
