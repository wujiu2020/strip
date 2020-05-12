package sessions

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (m *SessionManager) HasRemember(config *CookieConfig, r *http.Request) (psk, saltHash string, ok bool) {
	cfg := m.provider.Config()

	cookie, ok := getCookie(r.Cookies(), config.CookieRememberName)
	if !ok {
		return
	}

	value, createdAt, ok := DecodeSecureValue(cookie.Value, cfg.SecretKey)
	if !ok {
		return
	}

	if isExpired(createdAt, int2SecsDuration(config.RememberExpire)) {
		ok = false
		return
	}

	parts := strings.SplitN(value, COOKIE_VALUE_PARTS_SPLIT, 2)
	if len(parts) != 2 {
		return
	}

	uPsk, err := url.QueryUnescape(parts[0])
	if err != nil {
		ok = false
		return
	}
	uSaltHash, err := url.QueryUnescape(parts[1])
	if err != nil {
		ok = false
		return
	}

	psk = uPsk
	saltHash = uSaltHash
	return
}

func (m *SessionManager) ValidRemember(config *CookieConfig, psk, salt, saltHash string) (ok bool) {
	if psk == "" || salt == "" || saltHash == "" {
		return
	}

	cfg := m.provider.Config()

	uPsk, createdAt, ok := DecodeSecureValue(saltHash, cfg.SecretKey+salt)
	if !ok {
		return
	}

	if isExpired(createdAt, int2SecsDuration(config.RememberExpire)) {
		ok = false
		return
	}

	if uPsk != psk {
		ok = false
		return
	}

	return
}

func (m *SessionManager) WriteRemember(config *CookieConfig, w http.ResponseWriter, psk string, salt string) (ok bool) {
	cfg := m.provider.Config()

	createdAt := time.Now()

	value, ok := createRememberCookieValue(cfg.SecretKey, psk, salt, createdAt)
	if !ok {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     config.CookieRememberName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   config.CookieSecure,
		MaxAge:   config.RememberExpire,
	})
	return
}

func (m *SessionManager) DestroyRemember(config *CookieConfig, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     config.CookieRememberName,
		Path:     "/",
		HttpOnly: true,
		Secure:   config.CookieSecure,
		MaxAge:   -1,
	})
}

func createRememberCookieValue(sk, psk, salt string, createdAt time.Time) (hashValue string, ok bool) {
	if sk == "" || psk == "" || salt == "" {
		return
	}

	saltHash, ok := EncodeSecureValue(psk, sk+salt, createdAt)
	if !ok {
		return
	}

	value := url.QueryEscape(psk) + COOKIE_VALUE_PARTS_SPLIT + url.QueryEscape(saltHash)
	hashValue, ok = EncodeSecureValue(value, sk, createdAt)
	return
}
