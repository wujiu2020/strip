package sessions

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

type SessionManager struct {
	provider SessionProvider
}

func NewSessionManager(provider SessionProvider) *SessionManager {
	if provider.Config().SecretKey == "" {
		// SecretKey 为空可能导致安全问题，坚决 panic
		panic(ErrEmptySecretKey)
	}

	manager := new(SessionManager)
	manager.provider = provider
	return manager
}

func (m *SessionManager) GC(intervals ...time.Duration) {
	var interval time.Duration
	if len(intervals) > 0 {
		interval = intervals[0]

		// 小于1分钟的，默认回收时间设为1小时
		if interval < time.Minute {
			interval = time.Hour
		}
	}

	// 启动 Provider 的 GC
	err := m.provider.GC()
	if err != nil {
		log.Println("[SessionManager.GC] err:", err)
	}

	time.AfterFunc(interval, func() {
		// 时间到，再次执行 GC
		m.GC(interval)
	})
}

// 开始一个Session，从请求中获取Sid，或者创建一个新的
func (m *SessionManager) Start(config *CookieConfig, w http.ResponseWriter, r *http.Request) (sess SessionStore, createdAt time.Time, err error) {
	sid, createdAt, ok := m.ReadSidFromRequest(r, config.CookieName)
	if !ok {
		sess, err = m.createSession(CreateSid())
		if err != nil {
			return
		}

		createdAt = time.Now()
		m.WriteSessionCookie(config, r, w, sess.Sid(), createdAt)

	} else {
		sess, err = m.provider.Read(sid)
		if err == ErrNotFoundSession {
			sess, err = m.createSession(sid)
			if err != nil {
				return
			}
		}
	}
	return
}

// 删除当前的session
func (m *SessionManager) Destroy(config *CookieConfig, w http.ResponseWriter, r *http.Request) error {
	sid, _, ok := m.ReadSidFromRequest(r, config.CookieName)
	if !ok {
		return nil
	}

	// error can secure skip
	_ = m.provider.Destroy(sid)

	cookie := &http.Cookie{
		Name:     config.CookieName,
		Path:     "/",
		HttpOnly: true,
		Secure:   config.CookieSecure,
		Expires:  time.Now(),
		MaxAge:   -1,
		Domain:   config.CookieDomain,
	}

	http.SetCookie(w, cookie)

	return nil
}

// 为现有的session数据更换sid
func (m *SessionManager) Regenerate(config *CookieConfig, w http.ResponseWriter, r *http.Request, params ...map[string]interface{}) (sess SessionStore, createdAt time.Time, err error) {
	// 获取当前的sid，未找到时创建新的
	oldsid, _, ok := m.ReadSidFromRequest(r, config.CookieName)
	if !ok {
		sess, err = m.createSession(CreateSid(), params...)
		if err != nil {
			return
		}

		createdAt = time.Now()
		m.WriteSessionCookie(config, r, w, sess.Sid(), createdAt)
		return
	}

	sid := CreateSid()
	sess, err = m.provider.Regenerate(oldsid, sid)

	switch err {
	case ErrNotFoundSession:
		// 未找到时创建新的
		sess, err = m.createSession(sid, params...)
		if err != nil {
			return
		}
	}

	createdAt = time.Now()
	m.WriteSessionCookie(config, r, w, sess.Sid(), createdAt)
	return
}

// 设置 Session Cookie
func (m *SessionManager) WriteSessionCookie(config *CookieConfig, r *http.Request, w http.ResponseWriter, sid string, createdAt time.Time) {
	var (
		cfg = m.provider.Config()
	)

	// secure cookie value of sid
	value, ok := EncodeSecureValue(sid, cfg.SecretKey, createdAt)
	if !ok {
		return
	}

	cookie := &http.Cookie{
		Name:     config.CookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   config.CookieSecure,
		Domain:   config.CookieDomain,
	}

	if config.CookieExpire >= 0 {
		cookie.MaxAge = config.CookieExpire
	}

	http.SetCookie(w, cookie)
	r.AddCookie(cookie)

	return
}

// 创建新 Session
func (m *SessionManager) createSession(sid string, params ...map[string]interface{}) (sess SessionStore, err error) {
	sess, err = m.provider.Create(sid, params...)
	return
}

func (m *SessionManager) ReadSidFromRequest(r *http.Request, cookieName string) (sid string, created time.Time, ok bool) {
	var (
		config = m.provider.Config()
	)

	cookie, exists := getCookie(r.Cookies(), cookieName)
	if !exists {
		return
	}

	value, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return
	}

	value, vTime, exists := DecodeSecureValue(value, config.SecretKey)
	if !exists {
		return
	}

	// cookie 创建时间超过 session 失效时间，则判断为已过期
	if isExpired(vTime, int2SecsDuration(config.SessionExpire)) {
		return
	}

	sid = value
	created = vTime
	ok = true
	return
}
