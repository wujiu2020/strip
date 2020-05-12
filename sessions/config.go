package sessions

import (
	"net/http"
)

type CookieWriteHooker func(http.ResponseWriter, *http.Request, *http.Cookie)

type CookieConfig struct {
	CookieName         string // session cookie name
	CookieRememberName string // hashed value of user for auto login
	CookieSecure       bool   // is cookie use https?
	CookieDomain       string // session cookie domain
	CookieExpire       int    // session cookie expire seconds
	RememberExpire     int    // auto login remember expire seconds
}

type Config struct {
	SecretKey     string // secure secret key
	SessionExpire int    // session expire seconds
}
