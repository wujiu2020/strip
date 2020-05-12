package reqlogger

import (
	"net/http"
	"testing"

	"github.com/teapots/teapot"
)

func Test_RealIp(t *testing.T) {
	assert := &teapot.Assert{T: t}

	ip := "1.1.1.1"

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set(HeaderXRealIp, ip)

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set(HeaderXRealIp, " "+ip+" ")
	assert.True(realIp(req) == ip)

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set(HeaderXForwardedFor, ip+",0.0.0.0,2.2.2.2")
	assert.True(realIp(req) == ip)

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set(HeaderXForwardedFor, "  "+ip+" , 0.0.0.0, 2.2.2.2")
	assert.True(realIp(req) == ip)

	req, _ = http.NewRequest("GET", "/", nil)
	req.RemoteAddr = ip + ":"
	assert.True(realIp(req) == ip)

	req, _ = http.NewRequest("GET", "/", nil)
	req.RemoteAddr = ip + ":3000"
	assert.True(realIp(req) == ip)
}
