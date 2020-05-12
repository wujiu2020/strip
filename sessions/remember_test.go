package sessions

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_CreateRememberCookieValue(t *testing.T) {
	value, ok := createRememberCookieValue("sk", "psk", "salt", time.Now())
	assert.True(t, ok)
	assert.NotEmpty(t, value)

	value, ok = createRememberCookieValue("", "psk", "salt", time.Now())
	assert.False(t, ok)
	assert.Empty(t, value)

	value, ok = createRememberCookieValue("sk", "", "salt", time.Now())
	assert.False(t, ok)
	assert.Empty(t, value)

	value, ok = createRememberCookieValue("sk", "psk", "", time.Now())
	assert.False(t, ok)
	assert.Empty(t, value)
}

func Test_HasRemember(t *testing.T) {
	psk := "TEST_PSK"
	salt := "TEST_SALT"

	r := createRememberRequest(t, psk, salt, time.Now())

	uPsk, saltHash, ok := manager.HasRemember(cookieConfig, r)
	assert.True(t, ok)
	assert.Equal(t, psk, uPsk)

	r = createRememberRequest(t, psk, salt, time.Now().Add(time.Second*time.Duration(-cookieConfig.RememberExpire-100)))

	uPsk, saltHash, ok = manager.HasRemember(cookieConfig, r)
	assert.False(t, ok)
	assert.Empty(t, uPsk)
	assert.Empty(t, saltHash)
}

func Test_ValidRemember(t *testing.T) {
	psk := "TEST_PSK"
	salt := "TEST_SALT"

	r := createRememberRequest(t, psk, salt, time.Now())

	uPsk, saltHash, ok := manager.HasRemember(cookieConfig, r)
	assert.True(t, ok)
	assert.Equal(t, psk, uPsk)

	ok = manager.ValidRemember(cookieConfig, psk, salt, saltHash)
	assert.True(t, ok)

	ok = manager.ValidRemember(cookieConfig, psk, "", saltHash)
	assert.False(t, ok)

	ok = manager.ValidRemember(cookieConfig, "", salt, saltHash)
	assert.False(t, ok)

	ok = manager.ValidRemember(cookieConfig, psk, salt, "")
	assert.False(t, ok)

	ok = manager.ValidRemember(cookieConfig, "wrong", "wrong", "wrong")
	assert.False(t, ok)

	ok = manager.ValidRemember(cookieConfig, psk, salt, "wrong")
	assert.False(t, ok)
}

func createRememberRequest(t *testing.T, psk, salt string, createdAt time.Time) *http.Request {
	value, ok := createRememberCookieValue(config.SecretKey, psk, salt, createdAt)
	assert.True(t, ok)
	if !ok {
		t.FailNow()
	}

	cookie := &http.Cookie{
		Name:   cookieConfig.CookieRememberName,
		Value:  value,
		Path:   "/",
		Secure: true,
		MaxAge: cookieConfig.RememberExpire,
	}
	r, _ := http.NewRequest("GET", "https://www.qiniu.com", nil)
	r.AddCookie(cookie)

	return r
}
